package grpcclient

import (
	"fmt"
	"sync/atomic"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Pool struct {
	conns []*Client
	idx   atomic.Uint64
}

func NewPool(target string, size int, log *zap.Logger, opts ...Option) (*Pool, error) {
	if size < 1 {
		return nil, fmt.Errorf("grpcclient: pool size must be >= 1")
	}

	conns := make([]*Client, size)
	for i := range conns {
		c, err := New(target, log, opts...)
		if err != nil {
			for j := range i {
				conns[j].Close() //nolint:errcheck
			}
			return nil, fmt.Errorf("grpcclient: pool[%d]: %w", i, err)
		}
		conns[i] = c
	}

	log.Info("gRPC pool created", zap.String("target", target), zap.Int("size", size))
	return &Pool{conns: conns}, nil
}

func (p *Pool) Get() *grpc.ClientConn {
	n := p.idx.Add(1) - 1
	return p.conns[n%uint64(len(p.conns))].Conn()
}

func (p *Pool) Close() error {
	var lastErr error
	for _, c := range p.conns {
		if err := c.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

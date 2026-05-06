package grpcclient

import (
	"fmt"

	"github.com/gulmix/hermes/pkg/interceptor"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const roundRobinConfig = `{"loadBalancingConfig": [{"round_robin":{}}]}`

type Client struct {
	conn *grpc.ClientConn
}

func New(target string, log *zap.Logger, opts ...Option) (*Client, error) {
	cfg := defaultOptions()
	for _, o := range opts {
		o(&cfg)
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(cfg.tlsCreds),
		grpc.WithDefaultServiceConfig(roundRobinConfig),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithChainUnaryInterceptor(
			interceptor.MetadataUnary(),
			interceptor.TimeoutUnary(cfg.defaultTimeout),
			interceptor.RetryUnary(interceptor.RetryPolicy{
				MaxAttempts: cfg.maxRetries,
				Initial:     cfg.retryInitial,
				Max:         cfg.retryMax,
			}),
		),
		grpc.WithChainStreamInterceptor(
			interceptor.MetadataStream(),
		),
	}
	dialOpts = append(dialOpts, cfg.extraDialOpts...)

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("grpcclient: dial %s: %w", target, err)
	}

	log.Info("gRPC client created", zap.String("target", target))
	return &Client{conn: conn}, nil
}

func (c *Client) Conn() *grpc.ClientConn {
	return c.conn
}

func (c *Client) Close() error {
	return c.conn.Close()
}

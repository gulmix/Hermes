package grpcserver

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpc  *grpc.Server
	log   *zap.Logger
	addr  string
	isDev bool
}

func New(addr string, isDev bool, log *zap.Logger, opts ...grpc.ServerOption) *Server {
	s := grpc.NewServer(opts...)

	healthSvc := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthSvc)
	healthSvc.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	if isDev {
		reflection.Register(s)
	}

	return &Server{grpc: s, log: log, addr: addr, isDev: isDev}
}

func (s *Server) GRPC() *grpc.Server { return s.grpc }

func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.log.Info("gRPC server started", zap.String("addr", s.addr), zap.Bool("dev", s.isDev))

	errCh := make(chan error, 1)
	go func() { errCh <- s.grpc.Serve(lis) }()

	select {
	case <-ctx.Done():
		s.log.Info("shutting down gRPC server")
		s.grpc.GracefulStop()
		return nil
	case err := <-errCh:
		return err
	}
}

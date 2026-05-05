package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	userv1 "github.com/gulmix/hermes/gen/go/user/v1"
	"github.com/gulmix/hermes/services/user/internal/handler"
	"github.com/gulmix/hermes/services/user/internal/interceptor"
	"github.com/gulmix/hermes/services/user/internal/server"
	"github.com/gulmix/hermes/services/user/internal/telemetry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	otelShutdown, err := telemetry.Init(ctx, "user-service", "otel-collector:4317")
	if err != nil {
		log.Fatal("otel init", zap.Error(err))
	}
	defer otelShutdown(context.Background())

	isDev := os.Getenv("ENV") == "dev"
	opts := []grpc.ServerOption{
		interceptor.OTelStatsHandler(),
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryUnary(log),
			interceptor.LoggingUnary(log),
			interceptor.MetricsUnary(),
		),
		grpc.ChainStreamInterceptor(
			interceptor.RecoveryStream(log),
			interceptor.LoggingStream(log),
			interceptor.MetricsStream(),
		),
	}
	if !isDev {
		tlsCreds, err := server.LoadTLSCredentials(server.TLSConfig{
			CertFile:   "certs/server.crt",
			KeyFile:    "certs/server.key",
			CAFile:     "certs/ca.crt",
			ClientAuth: true,
		})
		if err != nil {
			log.Fatal("tls", zap.Error(err))
		}
		opts = append(opts, grpc.Creds(tlsCreds))
	}

	interceptor.ServeMetrics(":9090")

	srv := server.New(":50051", isDev, log, opts...)
	userv1.RegisterUserServiceServer(srv.GRPC(), handler.NewUserHandler(log))
	interceptor.InitializeMetrics(srv.GRPC())

	if err := srv.Run(ctx); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}

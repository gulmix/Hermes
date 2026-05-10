package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderv1 "github.com/gulmix/hermes/gen/go/order/v1"
	_ "github.com/gulmix/hermes/pkg/compression"
	"github.com/gulmix/hermes/pkg/grpcclient"
	"github.com/gulmix/hermes/pkg/grpcserver"
	"github.com/gulmix/hermes/pkg/interceptor"
	"github.com/gulmix/hermes/pkg/pprof"
	"github.com/gulmix/hermes/pkg/telemetry"
	"github.com/gulmix/hermes/services/order/internal/client"
	"github.com/gulmix/hermes/services/order/internal/handler"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	otelShutdown, err := telemetry.Init(ctx, "order-service", "otel-collector:4317")

	if err != nil {
		log.Fatal("otel init", zap.Error(err))
	}
	defer otelShutdown(context.Background())

	userTarget := envOr("USER_SERVICE_ADDR", "localhost:50051")

	userConn, err := grpcclient.New(userTarget, log,
		grpcclient.WithTimeout(3*time.Second),
		grpcclient.WithRetry(3, 100*time.Millisecond, 1*time.Second),
	)
	if err != nil {
		log.Fatal("user client", zap.Error(err))
	}
	defer userConn.Close()

	userClient := client.NewUserClient(userConn, log)

	if addr := os.Getenv("PPROF_ADDR"); addr != "" {
		go pprof.Serve(addr, log)
	}

	isDev := os.Getenv("ENV") == "dev"
	opts := grpcserver.DefaultKeepaliveOpts()
	opts = append(opts,
		interceptor.OTelStatsHandler(),
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryUnary(log),
			interceptor.LoggingUnary(log),
			interceptor.MetricsUnary(),
			interceptor.DeadlineUnary(),
		),
		grpc.ChainStreamInterceptor(
			interceptor.RecoveryStream(log),
			interceptor.LoggingStream(log),
			interceptor.MetricsStream(),
		),
	)

	interceptor.ServeMetrics(":9091")

	srv := grpcserver.New(":50052", isDev, log, opts...)
	orderv1.RegisterOrderServiceServer(srv.GRPC(), handler.NewOrderHandler(log, userClient))
	interceptor.InitializeMetrics(srv.GRPC())

	if err := srv.Run(ctx); err != nil {
		log.Fatal("server error", zap.Error(err))
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

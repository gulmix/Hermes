package interceptor

import (
	"net/http"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

var (
	Metrics = grpcprom.NewServerMetrics()
)

func MetricsUnary() grpc.UnaryServerInterceptor {
	return Metrics.UnaryServerInterceptor()
}

func MetricsStream() grpc.StreamServerInterceptor {
	return Metrics.StreamServerInterceptor()
}

func ServeMetrics(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(addr, mux)
}

func InitializeMetrics(s *grpc.Server) {
	Metrics.InitializeMetrics(s)
}

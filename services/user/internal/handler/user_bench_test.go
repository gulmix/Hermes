package handler_test

import (
	"context"
	"net"
	"testing"

	userv1 "github.com/gulmix/hermes/gen/go/user/v1"
	"github.com/gulmix/hermes/services/user/internal/handler"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func BenchmarkGetUser(b *testing.B) {
	h := handler.NewUserHandler(zap.NewNop())
	ctx := context.Background()
	req := &userv1.GetUserRequest{Id: "bench-id"}

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		_, _ = h.GetUser(ctx, req)
	}
}

func BenchmarkGetUserGRPC(b *testing.B) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		b.Fatal(err)
	}
	srv := grpc.NewServer()
	userv1.RegisterUserServiceServer(srv, handler.NewUserHandler(zap.NewNop()))
	go srv.Serve(lis) //nolint:errcheck
	defer srv.Stop()

	conn, err := grpc.NewClient(lis.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close()

	client := userv1.NewUserServiceClient(conn)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = client.GetUser(ctx, &userv1.GetUserRequest{Id: "bench-id"})
		}
	})
}

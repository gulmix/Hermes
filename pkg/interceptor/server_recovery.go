package interceptor

import (
	"context"
	"runtime/debug"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func recoveryHandler(log *zap.Logger) recovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p any) error {
		log.Error("panic recovered", zap.Any("panic", p), zap.ByteString("stack", debug.Stack()))
		return status.Errorf(codes.Internal, "internal server error")
	}
}

func RecoveryUnary(log *zap.Logger) grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor(recovery.WithRecoveryHandlerContext(recoveryHandler(log)))
}

func RecoveryStream(log *zap.Logger) grpc.StreamServerInterceptor {
	return recovery.StreamServerInterceptor(recovery.WithRecoveryHandlerContext(recoveryHandler(log)))
}

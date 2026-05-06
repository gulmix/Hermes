package interceptor

import (
	"context"

	grpcmw "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func zapLogger(log *zap.Logger) grpcmw.Logger {
	return grpcmw.LoggerFunc(func(ctx context.Context, lvl grpcmw.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)
		for i := 0; i+1 < len(fields); i += 2 {
			key, _ := fields[i].(string)
			f = append(f, zap.Any(key, fields[i+1]))
		}
		switch lvl {
		case grpcmw.LevelDebug:
			log.Debug(msg, f...)
		case grpcmw.LevelInfo:
			log.Info(msg, f...)
		case grpcmw.LevelWarn:
			log.Warn(msg, f...)
		case grpcmw.LevelError:
			log.Error(msg, f...)
		}
	})
}

func LoggingUnary(log *zap.Logger) grpc.UnaryServerInterceptor {
	return grpcmw.UnaryServerInterceptor(zapLogger(log))
}

func LoggingStream(log *zap.Logger) grpc.StreamServerInterceptor {
	return grpcmw.StreamServerInterceptor(zapLogger(log))
}

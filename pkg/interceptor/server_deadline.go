package interceptor

import (
	"context"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func DeadlineUnary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if dl, ok := ctx.Deadline(); ok {
			remaining := time.Until(dl)
			//nolint:errcheck
			grpc.SetHeader(ctx, metadata.Pairs(
				"x-deadline-remaining-ms", strconv.FormatInt(remaining.Milliseconds(), 10),
			))
		}
		return handler(ctx, req)
	}
}

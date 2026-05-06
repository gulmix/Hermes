package interceptor

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RetryPolicy struct {
	MaxAttempts int
	Initial     time.Duration
	Max         time.Duration
}

var retryableCodes = map[codes.Code]bool{
	codes.Unavailable:       true,
	codes.DeadlineExceeded:  true,
	codes.ResourceExhausted: true,
}

func RetryUnary(p RetryPolicy) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		backoff := p.Initial
		for attempt := 0; ; attempt++ {
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err == nil || attempt >= p.MaxAttempts || !isRetryable(err) {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				if backoff*2 < p.Max {
					backoff *= 2
				} else {
					backoff = p.Max
				}
			}
		}
	}
}

func isRetryable(err error) bool {
	s, ok := status.FromError(err)
	return ok && retryableCodes[s.Code()]
}

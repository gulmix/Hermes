package interceptor

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const headerRequestID = "x-request-id"

func MetadataUnary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(injectRequestID(ctx), method, req, reply, cc, opts...)
	}
}

func MetadataStream() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(injectRequestID(ctx), desc, cc, method, opts...)
	}
}

func injectRequestID(ctx context.Context) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}
	if len(md.Get(headerRequestID)) == 0 {
		md.Set(headerRequestID, uuid.NewString())
	}
	return metadata.NewOutgoingContext(ctx, md)
}

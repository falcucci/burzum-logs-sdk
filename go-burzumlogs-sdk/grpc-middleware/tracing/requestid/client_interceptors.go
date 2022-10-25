package grpc_requestid

import (
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/requestid"
)

// UnaryClientInterceptor returns a new unary client interceptor that adds request id  to the context.
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx = newClientRequestIDForCall(ctx)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamServerInterceptor returns a new streaming client interceptor that adds request id  to the context.
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx = newClientRequestIDForCall(ctx)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func newClientRequestIDForCall(ctx context.Context) context.Context {
	id := requestid.Extract(ctx)
	ctx = requestid.Inject(ctx, id)
	md := toMD(ctx, id)
	return metadata.NewOutgoingContext(ctx, md)
}

func toMD(ctx context.Context, requestID string) metadata.MD {

	md := metadata.Pairs(requestid.DefaultXRequestIDKey, requestID)

	if mdSource, ok := metadata.FromIncomingContext(ctx); ok {
		return metadata.Join(md, mdSource)
	}
	return md
}

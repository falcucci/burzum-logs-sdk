package grpc_requestid

import (
	"golang.org/x/net/context"

	"github.com/grpc-ecosystem/go-grpc-middleware"

	"google.golang.org/grpc"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/requestid"
	"google.golang.org/grpc/metadata"
)

// UnaryServerInterceptor returns a new unary server interceptors that adds request id to the context.
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := newRequestIDForCall(ctx)
		resp, err := handler(newCtx, req)
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds request id  to the context.
func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := newRequestIDForCall(stream.Context())
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx
		err := handler(srv, wrapped)

		return err
	}
}

func newRequestIDForCall(ctx context.Context) context.Context {

	md, ok := metadata.FromIncomingContext(ctx)
	var id string
	if ok {
		header, ok := md[requestid.DefaultXRequestIDKey]
		if ok && len(header) != 0 {
			id = header[0]
		}

	} else {
		id = requestid.Extract(ctx)
	}

	if id == "" {
		id = requestid.NewRequestID()
	}

	return requestid.Inject(ctx, id)
}

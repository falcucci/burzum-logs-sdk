package requestid

import (
	"golang.org/x/net/context"

	"github.com/renstrom/shortuuid"
)

// DefaultXRequestIDKey is metadata key name for request ID
var DefaultXRequestIDKey = "x-request-id"

type ctxMarker struct{}

var (
	ctxMarkerKey = &ctxMarker{}
)

func NewRequestID() string {
	return shortuuid.New()
}

// Extract takes the call-scoped requestID from context
func Extract(ctx context.Context) string {
	requestID, ok := ctx.Value(ctxMarkerKey).(string)
	if !ok {
		return NewRequestID()
	}
	return requestID
}

// Inject the request id
func Inject(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, requestID)
}

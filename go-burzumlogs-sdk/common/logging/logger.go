package bzlogging

import (
	"golang.org/x/net/context"

	"github.com/sirupsen/logrus"
)

type ctxMarker struct{}

var (
	// L is an alias for the the standard logger.
	L = logrus.NewEntry(logrus.StandardLogger())

	ctxMarkerKey = &ctxMarker{}
)

// Extract takes the call-scoped logrus.Entry from context
//
// If the logger wasn't used, a StandardLogger `logrus.Entry` is returned. This makes it safe to
// use regardless.
func Extract(ctx context.Context) *logrus.Entry {
	l, ok := ctx.Value(ctxMarkerKey).(*logrus.Entry)
	if !ok {
		return L
	}
	return l
}

//Inject returns a copy of parent in which the logger entry
func Inject(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, ctxMarkerKey, entry)

}

//Logger is an alias to Extract
func Logger(ctx context.Context) *logrus.Entry {
	return Extract(ctx)
}

//WithLogger returns a copy of parent in which the logger entry
func WithLogger(parent context.Context, entry *logrus.Entry) context.Context {
	return Inject(parent, entry)
}

// RequestIDTorequestIDField uses the requestID value to log the request request id.
func RequestIDTorequestIDField(requestID string) (key string, value interface{}) {
	return "requestID", requestID
}

func DefaultRequestIDfunc(ctx context.Context) (key string, value interface{}) {
	return RequestIDTorequestIDField("")
}

// RequestIDFromContext function determines how to extract the RequestID from Context
type RequestIDFromContext func(ctx context.Context) (key string, value interface{})

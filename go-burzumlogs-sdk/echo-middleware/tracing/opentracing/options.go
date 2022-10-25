package echo_opentracing

import (
	"github.com/labstack/echo/middleware"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
)

type options struct {
	skipper       middleware.Skipper
	tracer        opentracing.Tracer
	requestIDfunc bzlogging.RequestIDFromContext
}

var (
	defaultOptions = &options{
		skipper:       middleware.DefaultSkipper,
		tracer:        opentracing.GlobalTracer(),
		requestIDfunc: bzlogging.DefaultRequestIDfunc,
	}
)

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions

	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)

// WithSkipper customizes the function for skip the requests.
func WithSkipper(s middleware.Skipper) Option {
	return func(o *options) {
		o.skipper = s
	}
}

// WithTracer reference to the opentracing implementation.
func WithTracer(t opentracing.Tracer) Option {
	return func(o *options) {
		o.tracer = t
	}
}

// WithRequestId customizes the function for get the request id.
func WithRequestId(f bzlogging.RequestIDFromContext) Option {
	return func(o *options) {
		o.requestIDfunc = f
	}
}

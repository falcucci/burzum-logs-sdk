package echo_requestid

import (
	"github.com/labstack/echo/middleware"
)

type options struct {
	skipper middleware.Skipper
}

var (
	defaultOptions = &options{
		skipper: middleware.DefaultSkipper,
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

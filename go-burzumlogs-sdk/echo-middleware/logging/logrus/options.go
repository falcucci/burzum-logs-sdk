package echo_logzum

import (
	"github.com/labstack/echo/middleware"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
	"github.com/sirupsen/logrus"
)

var (
	defaultOptions = &options{
		levelFunc:     nil,
		skipper:       middleware.DefaultSkipper,
		requestIDfunc: bzlogging.DefaultRequestIDfunc,
	}
)

// ErrorToCode function determines the error code of an error
// This makes using custom errors with grpc middleware easier
type ErrorToCode func(err error) int

type options struct {
	levelFunc     CodeToLevel
	skipper       middleware.Skipper
	requestIDfunc bzlogging.RequestIDFromContext
}

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = DefaultCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
type CodeToLevel func(code int) logrus.Level

// WithRequestId customizes the function for get the request id.
func WithRequestId(f bzlogging.RequestIDFromContext) Option {
	return func(o *options) {
		o.requestIDfunc = f
	}
}

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.levelFunc = f
	}
}

// WithSkipper customizes the function for skip the requests.
func WithSkipper(s middleware.Skipper) Option {
	return func(o *options) {
		o.skipper = s
	}
}

// DefaultCodeToLevel is the default implementation of Echo return codes to log levels for server side.
func DefaultCodeToLevel(code int) logrus.Level {
	switch {
	case code >= 200 && code <= 299:
		return logrus.InfoLevel
	case code >= 300 && code <= 499:
		return logrus.WarnLevel
	default:
		return logrus.ErrorLevel
	}
}

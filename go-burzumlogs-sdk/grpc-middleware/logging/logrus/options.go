package grpc_logzum

import (
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
)

var (
	defaultOptions = &options{
		levelFunc:     nil,
		codeFunc:      DefaultErrorToCode,
		requestIDfunc: bzlogging.DefaultRequestIDfunc,
	}
)

// ErrorToCode function determines the error code of an error
// This makes using custom errors with grpc middleware easier
type ErrorToCode func(err error) codes.Code

func DefaultErrorToCode(err error) codes.Code {
	return grpc.Code(err)
}

type options struct {
	levelFunc     CodeToLevel
	codeFunc      ErrorToCode
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
type CodeToLevel func(code codes.Code) logrus.Level

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

// WithCodes customizes the function for mapping errors to error codes.
func WithCodes(f ErrorToCode) Option {
	return func(o *options) {
		o.codeFunc = f
	}
}

// DefaultCodeToLevel is the default implementation of gRPC return codes to log levels for server side.
func DefaultCodeToLevel(code codes.Code) logrus.Level {
	switch code {
	case codes.OK:
		return logrus.InfoLevel
	case codes.Canceled:
		return logrus.InfoLevel
	case codes.Unknown:
		return logrus.ErrorLevel
	case codes.InvalidArgument:
		return logrus.InfoLevel
	case codes.DeadlineExceeded:
		return logrus.WarnLevel
	case codes.NotFound:
		return logrus.InfoLevel
	case codes.AlreadyExists:
		return logrus.InfoLevel
	case codes.PermissionDenied:
		return logrus.WarnLevel
	case codes.Unauthenticated:
		return logrus.InfoLevel // unauthenticated requests can happen
	case codes.ResourceExhausted:
		return logrus.WarnLevel
	case codes.FailedPrecondition:
		return logrus.WarnLevel
	case codes.Aborted:
		return logrus.WarnLevel
	case codes.OutOfRange:
		return logrus.WarnLevel
	case codes.Unimplemented:
		return logrus.ErrorLevel
	case codes.Internal:
		return logrus.ErrorLevel
	case codes.Unavailable:
		return logrus.WarnLevel
	case codes.DataLoss:
		return logrus.ErrorLevel
	default:
		return logrus.ErrorLevel
	}
}

// DefaultClientCodeToLevel is the default implementation of gRPC return codes to log levels for client side.
func DefaultClientCodeToLevel(code codes.Code) logrus.Level {
	switch code {
	case codes.OK:
		return logrus.DebugLevel
	case codes.Canceled:
		return logrus.DebugLevel
	case codes.Unknown:
		return logrus.InfoLevel
	case codes.InvalidArgument:
		return logrus.DebugLevel
	case codes.DeadlineExceeded:
		return logrus.InfoLevel
	case codes.NotFound:
		return logrus.DebugLevel
	case codes.AlreadyExists:
		return logrus.DebugLevel
	case codes.PermissionDenied:
		return logrus.InfoLevel
	case codes.Unauthenticated:
		return logrus.InfoLevel // unauthenticated requests can happen
	case codes.ResourceExhausted:
		return logrus.DebugLevel
	case codes.FailedPrecondition:
		return logrus.DebugLevel
	case codes.Aborted:
		return logrus.DebugLevel
	case codes.OutOfRange:
		return logrus.DebugLevel
	case codes.Unimplemented:
		return logrus.WarnLevel
	case codes.Internal:
		return logrus.WarnLevel
	case codes.Unavailable:
		return logrus.WarnLevel
	case codes.DataLoss:
		return logrus.WarnLevel
	default:
		return logrus.InfoLevel
	}
}

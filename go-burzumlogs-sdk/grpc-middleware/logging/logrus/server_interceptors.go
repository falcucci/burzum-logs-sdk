package grpc_logzum

import (
	"path"
	"time"

	"golang.org/x/net/context"

	"github.com/sirupsen/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware"

	"google.golang.org/grpc"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
)

var (
	// KindField describes the log gield used to incicate whether this is a server or a client log statment.
	KindField = "span.kind"
)

// UnaryServerInterceptor returns a new unary server interceptors that adds logrus.Entry to the context.
func UnaryServerInterceptor(entry *logrus.Entry, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx := newLoggerForCall(ctx, o, entry, info.FullMethod)

		startTime := time.Now()
		resp, err := handler(newCtx, req)
		time := timeDiff(startTime)

		code := o.codeFunc(err)
		level := o.levelFunc(code)
		fields := logrus.Fields{
			"grpc.code":           code.String(),
			"grpc.duration":       time.Nanoseconds(),
			"grpc.duration_human": time.String(),
		}
		if err != nil {
			fields[logrus.ErrorKey] = err
		}
		levelLogf(
			bzlogging.Extract(newCtx).WithFields(fields), // re-extract logger from newCtx, as it may have extra fields that changed in the holder.
			level,
			"finished unary call")
		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds logrus.Entry to the context.
func StreamServerInterceptor(entry *logrus.Entry, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		newCtx := newLoggerForCall(stream.Context(), o, entry, info.FullMethod)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		startTime := time.Now()
		err := handler(srv, wrapped)
		time := timeDiff(startTime)
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		fields := logrus.Fields{
			"grpc.code":           code.String(),
			"grpc.duration":       time.Nanoseconds(),
			"grpc.duration_human": time.String(),
		}
		if err != nil {
			fields[logrus.ErrorKey] = err
		}
		levelLogf(
			bzlogging.Extract(newCtx).WithFields(fields), // re-extract logger from newCtx, as it may have extra fields that changed in the holder.
			level,
			"finished streaming call")
		return err
	}
}

func timeDiff(then time.Time) time.Duration {
	return time.Now().Sub(then)
}

func levelLogf(entry *logrus.Entry, level logrus.Level, format string, args ...interface{}) {
	switch level {
	case logrus.DebugLevel:
		entry.Debugf(format, args...)
	case logrus.InfoLevel:
		entry.Infof(format, args...)
	case logrus.WarnLevel:
		entry.Warningf(format, args...)
	case logrus.ErrorLevel:
		entry.Errorf(format, args...)
	case logrus.FatalLevel:
		entry.Fatalf(format, args...)
	case logrus.PanicLevel:
		entry.Panicf(format, args...)
	}
}

func newLoggerForCall(ctx context.Context, o *options, entry *logrus.Entry, fullMethodString string) context.Context {
	requestIDField, requestID := o.requestIDfunc(ctx)
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	callLog := entry.WithFields(
		logrus.Fields{
			KindField:      "server",
			"grpc.service": service,
			"grpc.method":  method,
			requestIDField: requestID,
		})

	return bzlogging.Inject(ctx, callLog)
}

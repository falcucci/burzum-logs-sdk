package echo_logzum

import (
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/sirupsen/logrus"
	"github.com/labstack/echo"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
)

var (
	// KindField describes the log gield used to incicate whether this is a server or a client log statment.
	KindField = "span.kind"
)

// Logger returns a middleware that logs HTTP requests.
func Logger(entry *logrus.Entry, opts ...Option) echo.MiddlewareFunc {
	o := evaluateServerOpt(opts)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if o.skipper(c) {
				return next(c)
			}
			req := c.Request()
			newCtx := newLoggerForHttpCall(req.Context(), entry, o, req.RequestURI, req.Method)
			c.SetRequest(req.WithContext(newCtx))

			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			duration := time.Now().Sub(start)

			level := o.levelFunc(res.Status)

			fields := logrus.Fields{}

			fields["http.remote_ip"] = c.RealIP()

			fields["http.host"] = req.Host

			p := req.URL.Path
			if p == "" {
				p = "/"
			}
			fields["http.path"] = p

			fields["http.referer"] = req.Referer()

			fields["http.user_agent"] = req.UserAgent()

			fields["http.status"] = res.Status

			fields["http.duration"] = duration.Nanoseconds()

			fields["http.duration_human"] = duration.String()

			cl := req.Header.Get(echo.HeaderContentLength)
			if cl == "" {
				cl = "0"
			}
			fields["http.bytes_in"] = cl

			fields["http.bytes_out"] = strconv.FormatInt(res.Size, 10)

			levelLogf(
				bzlogging.Extract(newCtx).WithFields(fields), // re-extract logger from newCtx, as it may have extra fields that changed in the holder.
				level,
				"finished http call")
			return
		}
	}
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

func newLoggerForHttpCall(ctx context.Context, entry *logrus.Entry, o *options, uri, method string) context.Context {

	requestIDField, requestID := o.requestIDfunc(ctx)

	callLog := entry.WithFields(
		logrus.Fields{
			KindField:      "server",
			"http.uri":     uri,
			"http.method":  method,
			requestIDField: requestID,
		})

	return bzlogging.Inject(ctx, callLog)
}

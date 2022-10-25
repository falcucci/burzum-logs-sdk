package echo_opentracing

import (
	"github.com/labstack/echo"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	ot_log "github.com/opentracing/opentracing-go/log"
)

var (
	echoTag = opentracing.Tag{string(ext.Component), "echo"}
)

func Tracer(opts ...Option) echo.MiddlewareFunc {
	o := evaluateServerOpt(opts)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if o.skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()

			ctx := req.Context()
			parentSpanContext, err := o.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
			if err != nil && err != opentracing.ErrSpanContextNotFound {
				//err
			}
			span := o.tracer.StartSpan(
				req.Method+req.RequestURI,
				// this is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
				ext.RPCServerOption(parentSpanContext),
				echoTag,
			)
			defer span.Finish()
			c.SetRequest(req.WithContext(opentracing.ContextWithSpan(ctx, span)))
			if err = next(c); err != nil {
				ext.Error.Set(span, true)
				span.LogFields(ot_log.Error(err))
				c.Error(err)

			}
			ext.HTTPMethod.Set(span, req.Method)
			status := uint16(res.Status)
			ext.HTTPStatusCode.Set(span, status)
			ext.HTTPUrl.Set(span, req.RequestURI)
			span.SetTag("http.referer", req.Referer())
			span.SetTag("http.user_agent", req.UserAgent())

			requestIDField, requestID := o.requestIDfunc(ctx)
			span.SetTag(requestIDField, requestID)

			return
		}
	}
}

package echo_requestid

import (
	"github.com/labstack/echo"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/requestid"
)

// RequestID returns a middleware that create or user requestid for HTTP requests.
func RequestID(opts ...Option) echo.MiddlewareFunc {
	o := evaluateServerOpt(opts)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {

			if o.skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			rid := req.Header.Get(requestid.DefaultXRequestIDKey)

			if rid == "" {
				rid = requestid.NewRequestID()
			}

			req.Header.Set(requestid.DefaultXRequestIDKey, rid)
			res.Header().Set(requestid.DefaultXRequestIDKey, rid)

			ctx := requestid.Inject(req.Context(), rid)

			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}

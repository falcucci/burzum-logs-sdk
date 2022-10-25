package echo_opentracing_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"

	"github.com/labstack/echo"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/requestid"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/echo-middleware/tracing/opentracing"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/echo-middleware/tracing/requestid"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
)

func TestLogger(t *testing.T) {

	tracer := mocktracer.New()

	opentracing.InitGlobalTracer(tracer)

	opts := []echo_opentracing.Option{
		echo_opentracing.WithRequestId(func(ctx context.Context) (string, interface{}) {
			return bzlogging.RequestIDTorequestIDField(requestid.Extract(ctx))
		}),
		echo_opentracing.WithTracer(tracer),
	}

	middlewares := []echo.MiddlewareFunc{
		echo_requestid.RequestID(),
		echo_opentracing.Tracer(opts...),
	}

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Header.Add("X-Request-ID", `foo`)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	// Status 200
	h(c)

	spans := tracer.FinishedSpans()

	assert.Equal(t, 1, len(spans))

	for _, span := range spans {
		assert.Equal(t, "foo", span.Tags()["requestID"], "all lines must contain `requestID`")
		assert.Equal(t, "/", span.Tags()["http.url"], "all lines must contain url name")
		assert.Equal(t, "GET", span.Tags()["http.method"], "all lines must contain method name")
		assert.Equal(t, uint16(200), span.Tags()["http.status_code"], "all lines must contain status")
	}

}

package echo_requestid_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"github.com/labstack/echo"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/requestid"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/echo-middleware/logging/logrus"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/echo-middleware/tracing/requestid"
)

func TestRequestID(t *testing.T) {
	logger, hook := test.NewNullLogger()

	entry := logrus.NewEntry(logger)

	opts := []echo_logzum.Option{
		echo_logzum.WithRequestId(func(ctx context.Context) (string, interface{}) {
			return bzlogging.RequestIDTorequestIDField(requestid.Extract(ctx))
		}),
	}

	middlewares := []echo.MiddlewareFunc{
		echo_requestid.RequestID(),
		echo_logzum.Logger(entry, opts...),
	}

	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Header.Add("X-Request-ID", `foo`)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	h := func(c echo.Context) error {
		bzlogging.Logger(c.Request().Context()).Info("test")
		return c.String(http.StatusOK, "test")
	}
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}

	// Status 200
	h(c)

	assert.Equal(t, 2, len(hook.Entries))

	for _, entry := range hook.Entries {
		assert.Equal(t, "foo", entry.Data["requestID"], "all lines must contain `requestID`")
		assert.Equal(t, "/", entry.Data["http.uri"], "all lines must contain uri name")
		assert.Equal(t, "GET", entry.Data["http.method"], "all lines must contain method name")

	}
	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Equal(t, "test", hook.Entries[0].Message)
	assert.Equal(t, "finished http call", hook.Entries[1].Message)
	hook.Reset()
	assert.Nil(t, hook.LastEntry())

	// Status 3xx
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = func(c echo.Context) error {
		bzlogging.Logger(c.Request().Context()).Warn("test")
		return c.String(http.StatusTemporaryRedirect, "test")
	}
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	h(c)

	assert.Equal(t, 2, len(hook.Entries))

	for _, entry := range hook.Entries {
		assert.Equal(t, "foo", entry.Data["requestID"], "all lines must contain `requestID`")
		assert.Equal(t, "/", entry.Data["http.uri"], "all lines must contain uri name")
		assert.Equal(t, "GET", entry.Data["http.method"], "all lines must contain method name")

	}
	assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	assert.Equal(t, "test", hook.Entries[0].Message)
	assert.Equal(t, "finished http call", hook.Entries[1].Message)
	hook.Reset()
	assert.Nil(t, hook.LastEntry())

	// Status 4xx
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = func(c echo.Context) error {
		bzlogging.Logger(c.Request().Context()).Warn("test")
		return c.String(http.StatusNotFound, "test")
	}
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	h(c)
	assert.Equal(t, 2, len(hook.Entries))

	for _, entry := range hook.Entries {
		assert.Equal(t, "foo", entry.Data["requestID"], "all lines must contain `requestID`")
		assert.Equal(t, "/", entry.Data["http.uri"], "all lines must contain uri name")
		assert.Equal(t, "GET", entry.Data["http.method"], "all lines must contain method name")

	}
	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Equal(t, "test", hook.Entries[0].Message)
	assert.Equal(t, "finished http call", hook.Entries[1].Message)
	hook.Reset()
	assert.Nil(t, hook.LastEntry())

	// Status 5xx with empty path
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	h = func(c echo.Context) error {
		bzlogging.Logger(c.Request().Context()).Error("test")
		return errors.New("error")
	}
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	h(c)
	assert.Equal(t, 2, len(hook.Entries))

	for _, entry := range hook.Entries {
		assert.Equal(t, "foo", entry.Data["requestID"], "all lines must contain `requestID`")
		assert.Equal(t, "/", entry.Data["http.uri"], "all lines must contain uri name")
		assert.Equal(t, "GET", entry.Data["http.method"], "all lines must contain method name")

	}
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(t, "test", hook.Entries[0].Message)
	assert.Equal(t, "finished http call", hook.Entries[1].Message)
	hook.Reset()
	assert.Nil(t, hook.LastEntry())

}

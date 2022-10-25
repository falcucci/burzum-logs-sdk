package grpc_logzum_test

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/grpc-ecosystem/go-grpc-middleware"

	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/grpc-middleware/logging/logrus"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/grpc-middleware/tracing/requestid"
)

func TestLogrusServerSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	opts := []grpc_logzum.Option{
		grpc_logzum.WithLevels(customCodeToLevel),
		grpc_logzum.WithRequestId(func(ctx context.Context) (string, interface{}) {
			return bzlogging.RequestIDTorequestIDField("foo")
		}),
	}
	b := newLogrusBaseSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_requestid.StreamServerInterceptor(),
			grpc_logzum.StreamServerInterceptor(logrus.NewEntry(b.logger), opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_requestid.UnaryServerInterceptor(),
			grpc_logzum.UnaryServerInterceptor(logrus.NewEntry(b.logger), opts...)),
	}
	suite.Run(t, &logrusServerSuite{b})
}

type logrusServerSuite struct {
	*logrusBaseSuite
}

func (s *logrusServerSuite) TestPingError_WithCustomLevels() {
	for _, tcase := range []struct {
		code  codes.Code
		level logrus.Level
		msg   string
	}{
		{
			code:  codes.Internal,
			level: logrus.ErrorLevel,
			msg:   "Internal must remap to ErrorLevel in DefaultCodeToLevel",
		},
		{
			code:  codes.NotFound,
			level: logrus.InfoLevel,
			msg:   "NotFound must remap to InfoLevel in DefaultCodeToLevel",
		},
		{
			code:  codes.FailedPrecondition,
			level: logrus.WarnLevel,
			msg:   "FailedPrecondition must remap to WarnLevel in DefaultCodeToLevel",
		},
		{
			code:  codes.Unauthenticated,
			level: logrus.ErrorLevel,
			msg:   "Unauthenticated is overwritten to ErrorLevel with customCodeToLevel override, which probably didn't work",
		},
	} {
		s.buffer.Reset()
		_, err := s.Client.PingError(
			s.SimpleCtx(),
			&pb_testproto.PingRequest{Value: "something", ErrorCodeReturned: uint32(tcase.code)})
		assert.Error(s.T(), err, "each call here must return an error")
		msgs := s.getOutputJSONs()
		require.Len(s.T(), msgs, 1, "only the interceptor log message is printed in PingErr")
		m := msgs[0]
		assert.Contains(s.T(), m, `"grpc.service": "mwitkow.testproto.TestService"`, "all lines must contain service name")
		assert.Contains(s.T(), m, `"grpc.method": "PingError"`, "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"grpc.code": "%s"`, tcase.code.String()), "all lines must contain method name")
		assert.Contains(s.T(), m, fmt.Sprintf(`"level": "%s"`, tcase.level.String()), tcase.msg)
	}
}

func (s *logrusServerSuite) TestPingList_WithRequestID() {
	stream, err := s.Client.PingList(s.SimpleCtx(), goodPing)
	require.NoError(s.T(), err, "should not fail on establishing the stream")
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		require.NoError(s.T(), err, "reading stream should not fail")
	}
	msgs := s.getOutputJSONs()
	assert.Len(s.T(), msgs, 2, "two log statements should be logged")
	for _, m := range msgs {
		s.T()
		assert.Contains(s.T(), m, `"grpc.service": "mwitkow.testproto.TestService"`, "all lines must contain service name")
		assert.Contains(s.T(), m, `"grpc.method": "PingList"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"requestID": "foo"`, "all lines must contain `requestID`")
	}
	assert.Contains(s.T(), msgs[0], `"msg": "some pinglist"`, "handler's message must contain user message")
	assert.Contains(s.T(), msgs[1], `"msg": "finished streaming call"`, "interceptor message must contain string")
	assert.Contains(s.T(), msgs[1], `"level": "info"`, "OK error codes must be logged on info level.")
	assert.Contains(s.T(), msgs[1], `"grpc.duration":`, "interceptor log statement should contain execution time")
	assert.Contains(s.T(), msgs[1], `"grpc.duration_human":`, "interceptor log statement should contain execution time")
}

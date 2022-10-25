package grpc_requestid_test

import (
	"runtime"
	"strings"
	"testing"

	"golang.org/x/net/context"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/grpc-middleware/tracing/requestid"

	"github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/requestid"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/grpc-middleware/logging/logrus"
)

func TestLogrusClientSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}

	b := newRequestIDBaseSuite(t)
	b.logger.Level = logrus.DebugLevel // a lot of our stuff is on debug level by default
	b.InterceptorTestSuite.ClientOpts = []grpc.DialOption{
		grpc.WithUnaryInterceptor(grpc_requestid.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(grpc_requestid.StreamClientInterceptor()),
	}

	opts := []grpc_logzum.Option{
		grpc_logzum.WithRequestId(func(ctx context.Context) (string, interface{}) {
			return bzlogging.RequestIDTorequestIDField(requestid.Extract(ctx))
		}),
	}
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_requestid.StreamServerInterceptor(),
			grpc_logzum.StreamServerInterceptor(logrus.NewEntry(b.logger), opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_requestid.UnaryServerInterceptor(),
			grpc_logzum.UnaryServerInterceptor(logrus.NewEntry(b.logger), opts...)),
	}
	suite.Run(t, &requestIDClientSuite{b})
}

type requestIDClientSuite struct {
	*requestIDBaseSuite
}

func (s *requestIDClientSuite) TestPingContext() {

	ctx := requestid.Inject(s.SimpleCtx(), "foo")

	_, err := s.Client.Ping(ctx, goodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	msgs := s.getOutputJSONs()

	assert.Len(s.T(), msgs, 2, "two log statements should be logged")
	for _, m := range msgs {
		s.T()
		assert.Contains(s.T(), m, `"grpc.service": "mwitkow.testproto.TestService"`, "all lines must contain service name")
		assert.Contains(s.T(), m, `"grpc.method": "Ping"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"requestID": "foo"`, "all lines must contain `requestID`")
	}
	assert.Contains(s.T(), msgs[0], `"msg": "some ping"`, "handler's message must contain user message")
	assert.Contains(s.T(), msgs[1], `"msg": "finished unary call"`, "interceptor message must contain string")
	assert.Contains(s.T(), msgs[1], `"level": "info"`, "OK error codes must be logged on info level.")
	assert.Contains(s.T(), msgs[1], `"grpc.duration":`, "interceptor log statement should contain execution time")
	assert.Contains(s.T(), msgs[1], `"grpc.duration_human":`, "interceptor log statement should contain execution time")
}

func (s *requestIDClientSuite) TestPing() {
	md := metadata.Pairs(requestid.DefaultXRequestIDKey, "foo")
	ctx := metadata.NewIncomingContext(s.SimpleCtx(), md)

	ctx = requestid.Inject(ctx, "foo")

	_, err := s.Client.Ping(ctx, goodPing)
	assert.NoError(s.T(), err, "there must be not be an on a successful call")
	msgs := s.getOutputJSONs()

	assert.Len(s.T(), msgs, 2, "two log statements should be logged")
	for _, m := range msgs {
		s.T()
		assert.Contains(s.T(), m, `"grpc.service": "mwitkow.testproto.TestService"`, "all lines must contain service name")
		assert.Contains(s.T(), m, `"grpc.method": "Ping"`, "all lines must contain method name")
		assert.Contains(s.T(), m, `"requestID": "foo"`, "all lines must contain `requestID`")
	}
	assert.Contains(s.T(), msgs[0], `"msg": "some ping"`, "handler's message must contain user message")
	assert.Contains(s.T(), msgs[1], `"msg": "finished unary call"`, "interceptor message must contain string")
	assert.Contains(s.T(), msgs[1], `"level": "info"`, "OK error codes must be logged on info level.")
	assert.Contains(s.T(), msgs[1], `"grpc.duration":`, "interceptor log statement should contain execution time")
	assert.Contains(s.T(), msgs[1], `"grpc.duration_human":`, "interceptor log statement should contain execution time")
}

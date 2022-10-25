package grpc_requestid_test

import (
	"io"
	"runtime"
	"strings"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/grpc-ecosystem/go-grpc-middleware"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/grpc-middleware/logging/logrus"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/requestid"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/grpc-middleware/tracing/requestid"

	"google.golang.org/grpc/metadata"
)

func TestRequestIDServerSuite(t *testing.T) {
	if strings.HasPrefix(runtime.Version(), "go1.7") {
		t.Skipf("Skipping due to json.RawMessage incompatibility with go1.7")
		return
	}
	opts := []grpc_logzum.Option{
		grpc_logzum.WithRequestId(func(ctx context.Context) (string, interface{}) {
			return bzlogging.RequestIDTorequestIDField(requestid.Extract(ctx))
		}),
	}
	b := newRequestIDBaseSuite(t)
	b.InterceptorTestSuite.ServerOpts = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_requestid.StreamServerInterceptor(),
			grpc_logzum.StreamServerInterceptor(logrus.NewEntry(b.logger), opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_requestid.UnaryServerInterceptor(),
			grpc_logzum.UnaryServerInterceptor(logrus.NewEntry(b.logger), opts...)),
	}
	suite.Run(t, &requestIDServerSuite{b})
}

type requestIDServerSuite struct {
	*requestIDBaseSuite
}

func (s *requestIDServerSuite) TestPingList_WithRequestIDFromMetadata() {

	md := metadata.Pairs(requestid.DefaultXRequestIDKey, "foo")
	ctx := metadata.NewOutgoingContext(s.SimpleCtx(), md)

	stream, err := s.Client.PingList(ctx, goodPing)
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

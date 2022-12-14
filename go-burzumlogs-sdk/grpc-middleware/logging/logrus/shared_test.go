package grpc_logzum_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"golang.org/x/net/context"

	"github.com/sirupsen/logrus"

	"github.com/grpc-ecosystem/go-grpc-middleware/testing"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/common/logging"
	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/grpc-middleware/logging/logrus"
	"google.golang.org/grpc/codes"
)

var (
	goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

type loggingPingService struct {
	pb_testproto.TestServiceServer
}

func (s *loggingPingService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	bzlogging.Extract(ctx).Info("some ping")
	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *loggingPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *loggingPingService) PingList(ping *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	bzlogging.Extract(stream.Context()).Info("some pinglist")
	return s.TestServiceServer.PingList(ping, stream)
}

func (s *loggingPingService) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

func customCodeToLevel(c codes.Code) logrus.Level {
	if c == codes.Unauthenticated {
		// Make this a special case for tests, and an error.
		return logrus.ErrorLevel
	}
	level := grpc_logzum.DefaultCodeToLevel(c)
	return level
}

type logrusBaseSuite struct {
	*grpc_testing.InterceptorTestSuite
	mutexBuffer *grpc_testing.MutexReadWriter
	buffer      *bytes.Buffer
	logger      *logrus.Logger
}

func newLogrusBaseSuite(t *testing.T) *logrusBaseSuite {
	b := &bytes.Buffer{}
	muB := grpc_testing.NewMutexReadWriter(b)
	logger := logrus.New()
	logger.Out = muB
	logger.Formatter = &logrus.JSONFormatter{DisableTimestamp: true}
	return &logrusBaseSuite{
		logger:      logger,
		buffer:      b,
		mutexBuffer: muB,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &loggingPingService{&grpc_testing.TestPingService{T: t}},
		},
	}
}

func (s *logrusBaseSuite) SetupTest() {
	s.mutexBuffer.Lock()
	s.buffer.Reset()
	s.mutexBuffer.Unlock()
}

func (s *logrusBaseSuite) getOutputJSONs() []string {
	ret := []string{}
	dec := json.NewDecoder(s.mutexBuffer)
	for {
		var val map[string]json.RawMessage
		err := dec.Decode(&val)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.T().Fatalf("failed decoding output from Logrus JSON: %v", err)
		}
		out, _ := json.MarshalIndent(val, "", "  ")
		ret = append(ret, string(out))
	}
	return ret
}

package grpc_requestid_test

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
)

var (
	goodPing = &pb_testproto.PingRequest{Value: "something", SleepTimeMs: 9999}
)

type requestIDPingService struct {
	pb_testproto.TestServiceServer
}

func (s *requestIDPingService) Ping(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
	bzlogging.Extract(ctx).Info("some ping")
	return s.TestServiceServer.Ping(ctx, ping)
}

func (s *requestIDPingService) PingError(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.Empty, error) {
	return s.TestServiceServer.PingError(ctx, ping)
}

func (s *requestIDPingService) PingList(ping *pb_testproto.PingRequest, stream pb_testproto.TestService_PingListServer) error {
	bzlogging.Extract(stream.Context()).Info("some pinglist")
	return s.TestServiceServer.PingList(ping, stream)
}

func (s *requestIDPingService) PingEmpty(ctx context.Context, empty *pb_testproto.Empty) (*pb_testproto.PingResponse, error) {
	return s.TestServiceServer.PingEmpty(ctx, empty)
}

type requestIDBaseSuite struct {
	*grpc_testing.InterceptorTestSuite
	mutexBuffer *grpc_testing.MutexReadWriter
	buffer      *bytes.Buffer
	logger      *logrus.Logger
}

func newRequestIDBaseSuite(t *testing.T) *requestIDBaseSuite {
	b := &bytes.Buffer{}
	muB := grpc_testing.NewMutexReadWriter(b)
	logger := logrus.New()
	logger.Out = muB
	logger.Formatter = &logrus.JSONFormatter{DisableTimestamp: true}
	return &requestIDBaseSuite{
		logger:      logger,
		buffer:      b,
		mutexBuffer: muB,
		InterceptorTestSuite: &grpc_testing.InterceptorTestSuite{
			TestService: &requestIDPingService{&grpc_testing.TestPingService{T: t}},
		},
	}
}

func (s *requestIDBaseSuite) SetupTest() {
	s.mutexBuffer.Lock()
	s.buffer.Reset()
	s.mutexBuffer.Unlock()
}

func (s *requestIDBaseSuite) getOutputJSONs() []string {
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

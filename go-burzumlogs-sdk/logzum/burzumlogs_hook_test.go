package logzum_test

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/luizalabs/burzumlogs-sdk/go-burzumlogs-sdk/logzum"
	"github.com/sirupsen/logrus"

	"net"

	"github.com/stretchr/testify/assert"
)

var (
	ln net.Listener
)

func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())

	var err error

	ln, err = net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Println(err)
		os.Exit(666)
	}
	defer ln.Close()

	os.Exit(m.Run())
}

func TestFire(t *testing.T) {
	var server net.Conn

	h, err := logzum.NewWithConfig("foo", logzum.Config{
		Host:            ln.Addr().String(),
		MaxRetries:      2,
		RetryDelay:      1 * time.Second,
		Buffersize:      1000,
		KeepAlivePeriod: 30 * time.Second,
		Formatter: &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "time",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		},
	})

	if err != nil {
		t.Error(err)
	}

	entry := &logrus.Entry{
		Message: "hello world!",
		Data:    logrus.Fields{"override": "yes"},
		Level:   logrus.DebugLevel,
	}
	if err := h.Fire(entry); err != nil {
		t.Error(err)
	}

	server, err = ln.Accept()

	defer server.Close()

	var res map[string]string
	if err := json.NewDecoder(server).Decode(&res); err != nil {
		t.Error(err)
		return
	}
	expected := map[string]string{
		"bztoken":  "foo",
		"time":     "0001-01-01T00:00:00Z",
		"level":    "debug",
		"message":  "hello world!",
		"override": "yes",
	}

	assert.EqualValues(t, expected, res)

}

func TestWithMinLevel(t *testing.T) {
	config := logzum.DefaultConfig
	config.MinLevel = logrus.InfoLevel

	h, err := logzum.NewWithConfig("foo", config)
	if err != nil {
		t.Error(err)
	}

	expected := []logrus.Level{
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	}

	assert.ObjectsAreEqualValues(expected, h.Levels())
}

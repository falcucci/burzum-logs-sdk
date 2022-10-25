package logzum

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

//Config holds the hook configs
type Config struct {
	Host            string        `yaml:"host"`
	MaxRetries      int           `yaml:"max-retries"`
	RetryDelay      time.Duration `yaml:"retry-delay"`
	KeepAlivePeriod time.Duration `yaml:"keep-alive"`
	Buffersize      int           `yaml:"buffer-size"`
	Formatter       logrus.Formatter
	Fields          map[string]interface{}
	MinLevel        logrus.Level
}

var (
	//DefaultConfig default configs
	DefaultConfig = Config{
		Host:            "tcp.burzum.appsluiza.com.br:5030",
		MaxRetries:      2,
		RetryDelay:      2 * time.Second,
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
		MinLevel: logrus.DebugLevel,
	}
)

type hook struct {
	mu sync.RWMutex

	conn io.WriteCloser

	entryC chan []byte

	done chan struct{}

	config Config

	bztoken string

	minLevel logrus.Level
}

//New create a new hook with default configs
func New(bztoken string) (logrus.Hook, error) {
	return NewWithConfig(bztoken, DefaultConfig)
}

//NewWithConfig create a new hook with custom configs
func NewWithConfig(bztoken string, config Config) (logrus.Hook, error) {

	if config.Host == "" {
		config.Host = DefaultConfig.Host
	}

	if config.Formatter == nil {
		config.Formatter = DefaultConfig.Formatter
	}

	bz := &hook{
		conn:     nil,
		entryC:   make(chan []byte, config.Buffersize),
		done:     make(chan struct{}),
		config:   config,
		bztoken:  bztoken,
		minLevel: config.MinLevel,
	}
	err := bz.connect()

	go bz.process()

	return bz, err

}

func (h *hook) Fire(entry *logrus.Entry) error {
	h.burzumFields(entry)
	serialized, err := h.config.Formatter.Format(entry)
	if err != nil {
		log.Printf("BurzumLogs: error on format %v\n", err)
		return err
	}

	select {
	case h.entryC <- serialized:
	default:
		log.Printf("BurzumLogs: sending buffer is full skipping messsage: %s", serialized)
	}
	return nil
}

func (h *hook) Levels() []logrus.Level {
	lvls := make([]logrus.Level, 0, len(logrus.AllLevels))
	for _, l := range logrus.AllLevels {
		if h.minLevel >= l {
			lvls = append(lvls, l)
		}
	}

	return lvls
}

func (h *hook) process() {
	for entry := range h.entryC {
		h.writeAndRetry(entry)
	}
	close(h.done)
}

func (h *hook) writeAndRetry(serialized []byte) {
	for i := 0; i < h.config.MaxRetries; i++ {
		if i > 0 {
			time.Sleep(h.config.RetryDelay)
		}

		if err := h.connect(); err != nil {
			log.Printf(err.Error())
			continue
		}

		n, err := h.conn.Write(serialized)
		if err != nil {
			log.Printf("Unable to send log line. Wrote %d bytes before error: %v\n", n, err)
			log.Printf("Making a new attempt for %s\n", serialized)
			h.conn.Close()
			h.conn = nil
			continue
		}

		break
	}

}

func (h *hook) Close() error {
	close(h.entryC)
	<-h.done
	h.conn.Close()
	return nil
}

func (h *hook) burzumFields(entry *logrus.Entry) {

	entry.Data["bztoken"] = h.bztoken

	for key, value := range h.config.Fields {
		entry.Data[key] = value
	}

}

func (h *hook) connect() error {
	defer h.mu.Unlock()
	h.mu.Lock()
	if h.conn == nil {
		tcpAddr, err := net.ResolveTCPAddr("tcp", h.config.Host)
		if err != nil {
			return fmt.Errorf("Unable to connect, error: %v", err)
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {

			h.conn = nil
			return fmt.Errorf("Unable to connect, error: %v", err)
		}
		conn.SetKeepAlive(true)
		conn.SetKeepAlivePeriod(h.config.KeepAlivePeriod)
		h.conn = conn

	}

	return nil
}

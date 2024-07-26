package kafkax

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/handler"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
	"go.uber.org/atomic"
)

const (
	bytesInMb        = 1024 * 1024
	defaultMsgSizeMb = 1
)

type Auth struct {
	Username string
	Password string
}

type ConsumerConfig struct {
	Brokers []string
	Topic   string
	GroupId string
	Auth    *Auth
}

// todo добавить middleware
func (c ConsumerConfig) DefaultConsumer(logger log.Logger, handler handler.SyncHandlerAdapter) Consumer {
	ctx := log.ToContext(context.Background(), log.String("topic", c.Topic))

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.Brokers,
		GroupID: c.GroupId,
		Topic:   c.Topic,
		Dialer: &kafka.Dialer{
			DualStack:     true,
			Timeout:       10 * time.Second,
			SASLMechanism: PlainAuth(c.Auth),
		},
		MinBytes:       1,
		MaxBytes:       64 * 1024 * 1024, //nolint:mnd
		CommitInterval: 1 * time.Second,
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			logger.Error(ctx, "kafka consumer: "+fmt.Sprintf(s, i...))
		}),
	})

	return Consumer{
		reader:    reader,
		handler:   handler,
		wg:        &sync.WaitGroup{},
		logger:    logger,
		close:     make(chan struct{}),
		alive:     atomic.NewBool(true),
		TopicName: c.Topic,
		observer:  NewLogObserver(ctx, logger),
	}
}

type PublisherConfig struct {
	Hosts            []string
	Topic            string
	MaxMsgSizeMb     int64
	WriteTimeoutSec  int
	RequiredAckLevel int
	Auth             *Auth
}

func (p PublisherConfig) DefaultPublisher(logger log.Logger, restMiddlewares ...publisher.Middleware) *publisher.Publisher {
	ctx := log.ToContext(context.Background(), log.String("topic", p.Topic))

	if p.MaxMsgSizeMb == 0 {
		logger.Debug(ctx, fmt.Sprintf("maxMsgSize is 0; set default maxMsgSize to %d Mb", defaultMsgSizeMb))
		p.MaxMsgSizeMb = defaultMsgSizeMb
	}

	writer := kafka.Writer{
		Addr:         kafka.TCP(p.Hosts...),
		Topic:        p.Topic,
		WriteTimeout: WithWriteTimeoutSecs(p.WriteTimeoutSec),
		RequiredAcks: WithRequiredAckLevel(p.RequiredAckLevel),
		BatchBytes:   p.MaxMsgSizeMb * bytesInMb,
		Transport: &kafka.Transport{
			SASL: PlainAuth(p.Auth),
		},
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			logger.Error(ctx, "kafka publisher: "+fmt.Sprintf(s, i...))
		}),
	}

	middlewares := []publisher.Middleware{
		publisher.PublisherRequestId(),
	}
	middlewares = append(middlewares, restMiddlewares...)

	pub := publisher.New(&writer, logger,
		NewLogObserver(ctx, logger),
		publisher.WithMiddlewares(middlewares...))

	pub.Topic = p.Topic
	pub.Address = writer.Addr.String()

	return pub
}

type Config struct {
	Publishers []*publisher.Publisher
	Consumers  []Consumer
}

func NewConfig(opts ...ConfigOption) Config {
	cfg := &Config{}

	for _, opt := range opts {
		opt(cfg)
	}

	return *cfg
}

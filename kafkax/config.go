package kafkax

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/kafka_metrics"
)

const (
	bytesInMb = 1024 * 1024
)

type Auth struct {
	Username string `validate:"required" schema:"Логин"`
	Password string `validate:"required" schema:"Пароль"`
}

type ConsumerConfig struct {
	Addresses         []string `validate:"required" schema:"Список адресов брокеров для чтения сообщений"`
	Topic             string   `validate:"required" schema:"Топик"`
	GroupId           string   `validate:"required" schema:"Идентификатор консьюмера"`
	Concurrency       int      `schema:"Кол-во обработчиков, по умолчанию 1"`
	MaxBatchSizeMb    int      `schema:"Максимальный размер батча для приема консьюмером, по умолчанию 64 Мб"`
	CommitIntervalSec *int     `schema:"Интервал в секундах с которым происходит коммит офсетов, по умолчанию 1 c"`
	Auth              *Auth    `schema:"Параметры аутентификации"`
}

func (c ConsumerConfig) DefaultConsumer(logger log.Logger, handler consumer.Handler,
	restMiddlewares ...consumer.Middleware) consumer.Consumer {
	ctx := log.ToContext(context.Background(), log.String("topic", c.Topic))

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.Addresses,
		GroupID: c.GroupId,
		Topic:   c.Topic,
		Dialer: &kafka.Dialer{
			DualStack:     true,
			Timeout:       10 * time.Second,
			SASLMechanism: PlainAuth(c.Auth),
		},
		MinBytes:       1,
		MaxBytes:       c.WithMaxBatchSize() * bytesInMb,
		CommitInterval: c.WithCommitIntervalSec(),
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			logger.Error(ctx, "kafka consumer: "+fmt.Sprintf(s, i...))
		}),
	})

	middlewares := []consumer.Middleware{
		ConsumerRequestId(),
	}
	middlewares = append(middlewares, restMiddlewares...)

	cons := consumer.New(
		logger,
		reader,
		handler,
		c.Concurrency,
		consumer.WithObserver(consumer.NewLogObserver(ctx, logger)),
		consumer.WithMiddlewares(middlewares...),
	)

	return *cons
}

type PublisherConfig struct {
	Addresses        []string `validate:"required" schema:"Список адресов брокеров для отправки сообщений"`
	Topic            string   `validate:"required" schema:"Топик для отправки сообщений описывается здесь либо в каждом сообщении"`
	MaxMsgSizeMb     int64    `schema:"Максимальный размер сообщений в Мб, по умолчанию 1 Мб"`
	BatchSize        int      `schema:"Количество буферизованных сообщений в пакетной отправке, по умолчанию 10"`
	BatchTimeoutMs   *int     `schema:"Периодичность записи батчей в кафку в мс, по умолчанию 500 мс"`
	WriteTimeoutSec  *int     `schema:"Таймаут отправки сообщений, по умолчанию 10 секунд"`
	RequiredAckLevel int      `schema:"Количество подтверждений реплик разделов для получения ответа на запрос отправки сообщения"`
	Auth             *Auth    `schema:"Параметры аутентификации"`
}

func (p PublisherConfig) DefaultPublisher(logger log.Logger, restMiddlewares ...publisher.Middleware) *publisher.Publisher {
	ctx := log.ToContext(context.Background(), log.String("topic", p.Topic))

	writer := kafka.Writer{
		Addr:         kafka.TCP(p.Addresses...),
		WriteTimeout: p.WithWriteTimeoutSecs(),
		RequiredAcks: p.WithRequiredAckLevel(),
		BatchBytes:   p.WithMaxMessageSize() * bytesInMb,
		BatchSize:    p.WithBatchSize(),
		BatchTimeout: p.WithBatchTimeoutMs(),
		Transport: &kafka.Transport{
			SASL: PlainAuth(p.Auth),
		},
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			logger.Error(ctx, "kafka publisher: "+fmt.Sprintf(s, i...))
		}),
	}

	middlewares := []publisher.Middleware{
		PublisherMetrics(kafka_metrics.NewPublisherStorage(metrics.DefaultRegistry)),
		PublisherRequestId(),
	}
	middlewares = append(middlewares, restMiddlewares...)

	pub := publisher.New(
		&writer,
		p.Topic,
		logger,
		publisher.WithObserver(publisher.NewLogObserver(ctx, logger, p.Topic)),
		publisher.WithMiddlewares(middlewares...),
	)

	return pub
}

type Config struct {
	Publishers []*publisher.Publisher
	Consumers  []consumer.Consumer
}

func NewConfig(opts ...ConfigOption) Config {
	cfg := &Config{}

	for _, opt := range opts {
		opt(cfg)
	}

	return *cfg
}

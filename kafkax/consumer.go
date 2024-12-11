package kafkax

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/log"
	"time"
)

type ConsumerConfig struct {
	Addresses         []string `validate:"required" schema:"Список адресов брокеров для чтения сообщений"`
	Topic             string   `validate:"required" schema:"Топик"`
	GroupId           string   `validate:"required" schema:"Идентификатор консьюмера"`
	Concurrency       int      `schema:"Кол-во обработчиков, по умолчанию 1"`
	MaxBatchSizeMb    int      `schema:"Максимальный размер батча для приема консьюмером, по умолчанию 64 Мб"`
	CommitIntervalSec *int     `schema:"Интервал в секундах с которым происходит коммит офсетов, по умолчанию 1 c"`
	Auth              *Auth    `schema:"Параметры аутентификации"`
	DialTimeoutMs     *int     `schema:"Таймаут установки соединения, по умолчанию 5 секунд"`
}

func (c ConsumerConfig) GetMaxBatchSizeMb() int {
	if c.MaxBatchSizeMb <= 0 {
		return 64
	}

	return c.MaxBatchSizeMb
}

func (c ConsumerConfig) GetCommitInterval() time.Duration {
	if c.CommitIntervalSec == nil {
		return 1 * time.Second
	}

	return time.Duration(*c.CommitIntervalSec) * time.Second
}

func (c ConsumerConfig) GetDialTimeout() time.Duration {
	if c.DialTimeoutMs == nil {
		return 5 * time.Second
	}
	return time.Duration(*c.DialTimeoutMs) * time.Millisecond
}

func (c ConsumerConfig) DefaultConsumer(
	logCtx context.Context,
	logger log.Logger,
	handler consumer.Handler,
	restMiddlewares ...consumer.Middleware,
) consumer.Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.Addresses,
		GroupID: c.GroupId,
		Topic:   c.Topic,
		Dialer: &kafka.Dialer{
			DualStack:     true,
			Timeout:       c.GetDialTimeout(),
			SASLMechanism: PlainAuth(c.Auth),
		},
		MinBytes:       1,
		MaxBytes:       c.GetMaxBatchSizeMb() * bytesInMb,
		CommitInterval: c.GetCommitInterval(),
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			logger.Error(logCtx, "kafka consumer: "+fmt.Sprintf(s, i...))
		}),
	})

	middlewares := []consumer.Middleware{
		ConsumerRequestId(),
	}
	middlewares = append(middlewares, restMiddlewares...)

	cons := consumer.New(
		reader,
		handler,
		c.Concurrency,
		consumer.WithObserver(consumer.NewLogObserver(logCtx, logger)),
		consumer.WithMiddlewares(middlewares...),
	)

	return *cons
}

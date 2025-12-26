package kafkax

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kprom"
	"github.com/txix-open/isp-kit/metrics"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/log"
)

const (
	defaultDialTimeout    = 5 * time.Second
	defaultMaxBatchSizeMb = 64
)

type ConsumerConfig struct {
	Addresses         []string `validate:"required" schema:"Список адресов брокеров для чтения сообщений"`
	Topic             string   `validate:"required" schema:"Топик"`
	GroupId           string   `validate:"required" schema:"Идентификатор консьюмера"`
	Concurrency       int      `schema:"Кол-во обработчиков, по умолчанию 1"`
	MaxBatchSizeMb    int32    `schema:"Максимальный размер батча для приема консьюмером, по умолчанию 64 Мб"`
	CommitIntervalSec *int     `schema:"Интервал в секундах с которым происходит коммит офсетов, по умолчанию 1 c"`
	Auth              *Auth    `schema:"Параметры аутентификации"`
	TLS               *TLS     `schema:"Данные для установки TLS-соединения"`
	DialTimeoutMs     *int     `schema:"Таймаут установки соединения, по умолчанию 5 секунд"`
	MetricConsumerId  *string  `schema:"Идентификатор консьюмера в метриках, при отсутствии метрики не отправляются"`
}

func (c ConsumerConfig) GetMaxBatchSizeMb() int32 {
	if c.MaxBatchSizeMb <= 0 {
		return defaultMaxBatchSizeMb
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
		return defaultDialTimeout
	}
	return time.Duration(*c.DialTimeoutMs) * time.Millisecond
}

func (c ConsumerConfig) DefaultConsumer(
	logCtx context.Context,
	logger log.Logger,
	handler consumer.Handler,
	restMiddlewares ...consumer.Middleware,
) consumer.Consumer {
	var authMechanism string
	if c.Auth.Mechanism == nil {
		authMechanism = AuthTypePlain
	} else {
		authMechanism = *c.Auth.Mechanism
	}

	saslMechanism, err := setupSASLMechanism(authMechanism, c.Auth)
	if err != nil {
		logger.Error(logCtx, errors.WithMessage(err, "failed to setup sasl mechanism"))
	}

	tls, err := setupTLS(c.TLS)
	if err != nil {
		logger.Error(logCtx, errors.WithMessage(err, "failed to setup tls"))
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(c.Addresses...),
		kgo.ConsumerGroup(c.GroupId),
		kgo.DisableIdempotentWrite(),
		kgo.ConsumeTopics(c.Topic),
		kgo.DialTLSConfig(tls),
		kgo.DialTimeout(c.GetDialTimeout()),
		kgo.SASL(saslMechanism),
		kgo.FetchMinBytes(1),
		kgo.FetchMaxBytes(c.GetMaxBatchSizeMb() * bytesInMb),
		kgo.AutoCommitInterval(c.GetCommitInterval()),
		kgo.WithLogger(NewLogger(logCtx, "kafka consumer", kgo.LogLevelError, logger)),
	}

	middlewares := []consumer.Middleware{
		ConsumerRequestId(),
	}
	middlewares = append(middlewares, restMiddlewares...)

	if c.MetricConsumerId != nil {
		labels := prometheus.Labels{
			"consumerId": *c.MetricConsumerId,
		}
		consumerMetrics := kprom.NewMetrics(
			"kafka_consumer",
			kprom.WithStaticLabel(labels),
		)
		opts = append(opts, kgo.WithHooks(consumerMetrics))
		defaultRegistry := metrics.DefaultRegistry
		defaultRegistry.GetOrRegister(consumerMetrics)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		logger.Error(logCtx, errors.WithMessage(err, "create kafka client"))
	}

	cons := consumer.New(
		client,
		c.GroupId,
		handler,
		c.Concurrency,
		consumer.WithObserver(consumer.NewLogObserver(logCtx, logger)),
		consumer.WithMiddlewares(middlewares...),
	)

	return *cons
}

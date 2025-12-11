package kafkax

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kprom"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/kafka_metrics"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
)

const (
	defaultWriteTimeoutSec       = 10
	defaultMaxMsgSizeMb          = 64
	defaultBatchSizePerPartition = 10
	defaultBatchTimeoutMs        = 500
	defaultDialTimeoutSec        = 5
)

type PublisherConfig struct {
	Addresses                  []string `validate:"required" schema:"Список адресов брокеров для отправки сообщений"`
	Topic                      string   `validate:"required" schema:"Топик для отправки сообщений описывается здесь либо в каждом сообщении"`
	MaxMsgSizeMbPerPartition   int32    `schema:"Максимальный размер сообщений в Мб, по умолчанию 64 Мб"`
	BatchSizePerPartition      int      `schema:"Количество буферизованных сообщений в пакетной отправке, по умолчанию 10"`
	BatchTimeoutPerPartitionMs *int     `schema:"Периодичность записи батчей в кафку в мс, по умолчанию 500 мс"`
	WriteTimeoutSec            *int     `schema:"Таймаут отправки сообщений, по умолчанию 10 секунд"`
	RequiredAckLevel           int      `schema:"Количество подтверждений реплик разделов для получения ответа на запрос отправки сообщения"`
	Auth                       *Auth    `schema:"Параметры аутентификации"`
	TLS                        *TLS     `schema:"Данные для установки TLS-соединения"`
	DialTimeoutMs              *int     `schema:"Таймаут установки соединения, по умолчанию 5 секунд"`
	MetricPublisherId          *string  `schema:"Идентификатор паблишера в метриках, при отсутствии метрики не отправляются"`
}

func (p PublisherConfig) GetRequiredAckLevel() kgo.Acks {
	switch p.RequiredAckLevel {
	case 1:
		return kgo.LeaderAck()
	case -1:
		return kgo.AllISRAcks()
	}

	return kgo.NoAck()
}

func (p PublisherConfig) GetWriteTimeout() time.Duration {
	if p.WriteTimeoutSec == nil {
		return time.Duration(defaultWriteTimeoutSec) * time.Second
	}
	return time.Duration(*p.WriteTimeoutSec) * time.Second
}

func (p PublisherConfig) GetMaxMessageSizePerPartition() int32 {
	if p.MaxMsgSizeMbPerPartition <= 0 {
		return defaultMaxMsgSizeMb
	}
	return p.MaxMsgSizeMbPerPartition
}

func (p PublisherConfig) GetBatchSizePerPartition() int {
	if p.BatchSizePerPartition <= 0 {
		return defaultBatchSizePerPartition
	}
	return p.BatchSizePerPartition
}

func (p PublisherConfig) GetBatchTimeoutPerPartition() time.Duration {
	if p.BatchTimeoutPerPartitionMs == nil {
		return time.Duration(defaultBatchTimeoutMs) * time.Millisecond
	}
	return time.Duration(*p.BatchTimeoutPerPartitionMs) * time.Millisecond
}

func (p PublisherConfig) GetDialTimeout() time.Duration {
	if p.DialTimeoutMs == nil {
		return time.Duration(defaultDialTimeoutSec) * time.Second
	}
	return time.Duration(*p.DialTimeoutMs) * time.Millisecond
}

func (p PublisherConfig) DefaultPublisher(
	logCtx context.Context,
	logger log.Logger,
	restMiddlewares ...publisher.Middleware,
) *publisher.Publisher {
	var authMechanism string
	if p.Auth.Mechanism == nil {
		authMechanism = AuthTypePlain
	} else {
		authMechanism = *p.Auth.Mechanism
	}

	saslMechanism, err := setupSASLMechanism(authMechanism, p.Auth)
	if err != nil {
		logger.Error(logCtx, errors.WithMessage(err, "failed to setup sasl mechanism"))
	}

	tls, err := setupTLS(p.TLS)
	if err != nil {
		logger.Error(logCtx, errors.WithMessage(err, "failed to setup tls"))
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(p.Addresses...),
		kgo.ProduceRequestTimeout(p.GetWriteTimeout()),
		kgo.DisableIdempotentWrite(),
		kgo.RequiredAcks(p.GetRequiredAckLevel()),
		kgo.ProducerBatchMaxBytes(p.GetMaxMessageSizePerPartition() * bytesInMb),
		kgo.MaxBufferedRecords(p.GetBatchSizePerPartition()),
		kgo.ProducerLinger(p.GetBatchTimeoutPerPartition()),
		kgo.DialTimeout(p.GetDialTimeout()),
		kgo.SASL(saslMechanism),
		kgo.DialTLSConfig(tls),
		kgo.WithLogger(NewLogger(logCtx, "kafka publisher", kgo.LogLevelError, logger)),
	}

	middlewares := []publisher.Middleware{
		PublisherMetrics(kafka_metrics.NewPublisherStorage(metrics.DefaultRegistry)),
		PublisherRequestId(),
	}
	middlewares = append(middlewares, restMiddlewares...)

	if p.MetricPublisherId != nil {
		labels := prometheus.Labels{
			"publisherId": *p.MetricPublisherId,
		}
		publisherMetrics := kprom.NewMetrics(
			"kafka_publisher",
			kprom.WithStaticLabel(labels),
		)
		opts = append(opts, kgo.WithHooks(publisherMetrics))
		defaultRegistry := metrics.DefaultRegistry
		defaultRegistry.GetOrRegister(publisherMetrics)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		logger.Error(logCtx, errors.WithMessage(err, "create kafka client"))
	}

	pub := publisher.New(
		client,
		p.Topic,
		publisher.WithMiddlewares(middlewares...),
	)

	return pub
}

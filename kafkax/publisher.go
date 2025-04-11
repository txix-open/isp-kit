package kafkax

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/kafka_metrics"
)

type PublisherConfig struct {
	Addresses                  []string `validate:"required" schema:"Список адресов брокеров для отправки сообщений"`
	Topic                      string   `validate:"required" schema:"Топик для отправки сообщений описывается здесь либо в каждом сообщении"`
	MaxMsgSizeMbPerPartition   int64    `schema:"Максимальный размер сообщений в Мб, по умолчанию 64 Мб"`
	BatchSizePerPartition      int      `schema:"Количество буферизованных сообщений в пакетной отправке, по умолчанию 10"`
	BatchTimeoutPerPartitionMs *int     `schema:"Периодичность записи батчей в кафку в мс, по умолчанию 500 мс"`
	WriteTimeoutSec            *int     `schema:"Таймаут отправки сообщений, по умолчанию 10 секунд"`
	RequiredAckLevel           int      `schema:"Количество подтверждений реплик разделов для получения ответа на запрос отправки сообщения"`
	Auth                       *Auth    `schema:"Параметры аутентификации"`
	TLS                        *TLS     `schema:"Данные для установки TLS-соединения"`
	DialTimeoutMs              *int     `schema:"Таймаут установки соединения, по умолчанию 5 секунд"`
	PublisherMetricId          *string  `schema:"Идентификатор паблишера в метриках, при отсутствии метрики не отправляются"`
}

func (p PublisherConfig) GetWriteTimeout() time.Duration {
	if p.WriteTimeoutSec == nil {
		return 10 * time.Second
	}

	return time.Duration(*p.WriteTimeoutSec) * time.Second
}

func (p PublisherConfig) GetRequiredAckLevel() kafka.RequiredAcks {
	if p.RequiredAckLevel <= 1 && p.RequiredAckLevel >= -1 {
		return kafka.RequiredAcks(p.RequiredAckLevel)
	}

	return kafka.RequireNone
}

func (p PublisherConfig) GetMaxMessageSizePerPartition() int64 {
	if p.MaxMsgSizeMbPerPartition <= 0 {
		return 64
	}

	return p.MaxMsgSizeMbPerPartition
}

func (p PublisherConfig) GetBatchSizePerPartition() int {
	if p.BatchSizePerPartition <= 0 {
		return 10
	}

	return p.BatchSizePerPartition
}

func (p PublisherConfig) GetBatchTimeoutPerPartition() time.Duration {
	if p.BatchTimeoutPerPartitionMs == nil {
		return 500 * time.Millisecond
	}

	return time.Duration(*p.BatchTimeoutPerPartitionMs) * time.Millisecond
}

func (p PublisherConfig) GetDialTimeout() time.Duration {
	if p.DialTimeoutMs == nil {
		return 5 * time.Second
	}
	return time.Duration(*p.DialTimeoutMs) * time.Millisecond
}

func (p PublisherConfig) createTransport(mechanismType string) (*kafka.Transport, error) {
	saslMechanism, err := setupSASLMechanism(mechanismType, p.Auth)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to setup sasl mechanism")
	}

	tls, err := setupTLS(p.TLS)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to setup tls")
	}

	transport := &kafka.Transport{
		DialTimeout: p.GetDialTimeout(),
		SASL:        saslMechanism,
		TLS:         tls,
	}

	return transport, nil
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

	transport, err := p.createTransport(authMechanism)
	if err != nil {
		logger.Error(logCtx, errors.WithMessage(err, "create kafka publisher transport"))
	}

	writer := kafka.Writer{
		Addr:         kafka.TCP(p.Addresses...),
		WriteTimeout: p.GetWriteTimeout(),
		RequiredAcks: p.GetRequiredAckLevel(),
		BatchBytes:   p.GetMaxMessageSizePerPartition() * bytesInMb,
		BatchSize:    p.GetBatchSizePerPartition(),
		BatchTimeout: p.GetBatchTimeoutPerPartition(),
		Transport:    transport,
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			logger.Error(logCtx, "kafka publisher: "+fmt.Sprintf(s, i...))
		}),
	}

	middlewares := []publisher.Middleware{
		PublisherMetrics(kafka_metrics.NewPublisherStorage(metrics.DefaultRegistry)),
		PublisherRequestId(),
	}
	middlewares = append(middlewares, restMiddlewares...)

	isSendMetrics := false
	if p.PublisherMetricId != nil {
		isSendMetrics = true
	}

	pub := publisher.New(
		&writer,
		p.Topic,
		publisher.NewMetrics(isSendMetrics, sendMetricPeriod),
		publisher.WithMiddlewares(middlewares...),
	)

	return pub
}

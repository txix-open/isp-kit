package kafkax

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/log"
)

type ConsumerConfig struct {
	Addresses         []string `validate:"required" schema:"Список адресов брокеров для чтения сообщений"`
	Topic             string   `validate:"required" schema:"Топик"`
	GroupId           string   `validate:"required" schema:"Идентификатор консьюмера"`
	Concurrency       int      `schema:"Кол-во обработчиков, по умолчанию 1"`
	MaxBatchSizeMb    int      `schema:"Максимальный размер батча для приема консьюмером, по умолчанию 64 Мб"`
	CommitIntervalSec *int     `schema:"Интервал в секундах с которым происходит коммит офсетов, по умолчанию 1 c"`
	Auth              *Auth    `schema:"Параметры аутентификации"`
	TLS               *TLS     `schema:"Данные для установки TLS-соединения"`
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

func (c ConsumerConfig) createDialer(mechanismType string) (*kafka.Dialer, error) {
	var saslMechanism sasl.Mechanism

	switch mechanismType {
	case AuthTypePlain:
		saslMechanism = PlainAuth(c.Auth)
	case AuthTypeSCRAM:
		var err error
		saslMechanism, err = ScramAuth(c.Auth)
		if err != nil {
			return nil, errors.WithMessage(err, "get scram auth mechanism")
		}
	default:
		return nil, errors.Errorf("unexpected sasl mechanism: %s", mechanismType)
	}

	dialer := &kafka.Dialer{
		DualStack:     true,
		Timeout:       c.GetDialTimeout(),
		SASLMechanism: saslMechanism,
	}

	if c.TLS != nil {
		rawCert, err := base64.StdEncoding.DecodeString(c.TLS.Certificate)
		if err != nil {
			return nil, errors.WithMessage(err, "decode base64 tls certificate")
		}

		dialer.TLS = &tls.Config{
			Certificates: []tls.Certificate{
				{
					Certificate: [][]byte{rawCert},
					PrivateKey:  c.TLS.PrivateKey,
				},
			},
		}
	}

	return dialer, nil
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

	dialer, err := c.createDialer(authMechanism)
	if err != nil {
		logger.Error(logCtx, errors.WithMessage(err, "create kafka consumer dialer"))
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.Addresses,
		GroupID:        c.GroupId,
		Topic:          c.Topic,
		Dialer:         dialer,
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

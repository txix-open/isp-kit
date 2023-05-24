package grmqx

import (
	"fmt"
	"net/url"
	"time"

	"github.com/integration-system/grmq/consumer"
	"github.com/integration-system/grmq/publisher"
	"github.com/integration-system/grmq/retry"
	"github.com/integration-system/grmq/topology"
	"github.com/integration-system/isp-kit/metrics"
	rabbitmq_metircs "github.com/integration-system/isp-kit/metrics/rabbitmq_metrics"
	"github.com/rabbitmq/amqp091-go"
)

type Connection struct {
	Host     string `valid:"required" schema:"Хост"`
	Port     int    `valid:"required" schema:"Порт"`
	Username string `schema:"Логин"`
	Password string `schema:"Пароль"`
	Vhost    string `schema:"Виртуальный хост"`
}

func (c Connection) Url() string {
	u := url.URL{
		Scheme: "amqp",
		User:   nil,
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.Vhost,
	}
	if c.Username != "" {
		u.User = url.UserPassword(c.Username, c.Password)
	}
	return u.String()
}

type Publisher struct {
	Exchange   string `schema:"Точка обмена"`
	RoutingKey string `valid:"required" schema:"Ключ маршрутизации,для публикации напрямую в очередь указывается название очереди"`
}

func (p Publisher) DefaultPublisher(restMiddlewares ...publisher.Middleware) *publisher.Publisher {
	middlewares := append(
		[]publisher.Middleware{
			publisher.PersistentMode(),
			PublisherRequestId(),
			PublisherMetrics(rabbitmq_metircs.NewPublisherStorage(metrics.DefaultRegistry)),
		},
		restMiddlewares...,
	)
	return publisher.New(
		p.Exchange,
		p.RoutingKey,
		publisher.WithMiddlewares(middlewares...),
	)
}

type RetryConfig struct {
	DelayInMs   int `valid:"required" schema:"Задержка в миллисекундах"`
	MaxAttempts int `valid:"required" schema:"Количество попыток,-1 = бесконечно"`
}

type RetryPolicy struct {
	FinallyMoveToDlq bool          `schema:"Отправить в DLQ,отправить сообщение в DLQ в случае последней неудавшеймся попытки обработки"`
	Retries          []RetryConfig `schema:"Настройки"`
}

type Consumer struct {
	Queue              string       `valid:"required" schema:"Наименование очереди"`
	Dlq                bool         `schema:"Создать очередь DLQ"`
	PrefetchCount      int          `schema:"Количество предзагруженных сообщений,по умолчанию - 1"`
	Concurrency        int          `schema:"Количество обработчиков,по умолчанию - 1, рекомендовано использовать значение = prefetchCount"`
	DisableAutoDeclare bool         `schema:"Отключить автоматическое объявление,по умолчанию  exchange, queue и binding будут созданы автоматически"`
	Binding            *Binding     `schema:"Настройки топологии"`
	RetryPolicy        *RetryPolicy `schema:"Политика повторной обработки"`
}

type Binding struct {
	Exchange     string `valid:"required" schema:"Точка обмена"`
	ExchangeType string `valid:"required,in(direct|fanout|topic)" schema:"Тип точки обмена"`
	RoutingKey   string `valid:"required" schema:"Ключ маршрутизации"`
}

func (c Consumer) DefaultConsumer(handler consumer.Handler, restMiddlewares ...consumer.Middleware) consumer.Consumer {
	prefetchCount := c.PrefetchCount
	if prefetchCount <= 0 {
		prefetchCount = 1
	}
	concurrency := c.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	middlewares := append(
		[]consumer.Middleware{
			ConsumerRequestId(),
		},
		restMiddlewares...,
	)
	opts := []consumer.Option{
		consumer.WithPrefetchCount(prefetchCount),
		consumer.WithConcurrency(concurrency),
		consumer.WithMiddlewares(middlewares...),
	}
	if c.RetryPolicy != nil {
		policy := retryPolicyFromConfig(*c.RetryPolicy)
		opts = append(opts, consumer.WithRetryPolicy(policy))
	}
	return consumer.New(
		handler,
		c.Queue,
		opts...,
	)
}

func TopologyFromConsumers(consumers ...Consumer) topology.Declarations {
	opts := make([]topology.DeclarationsOption, 0)
	for _, consumer := range consumers {
		if consumer.DisableAutoDeclare {
			continue
		}

		queueOpts := []topology.QueueOption{
			topology.WithDLQ(consumer.Dlq),
		}
		if consumer.RetryPolicy != nil {
			policy := retryPolicyFromConfig(*consumer.RetryPolicy)
			queueOpts = append(queueOpts, topology.WithRetryPolicy(policy))
		}

		opts = append(opts, topology.WithQueue(consumer.Queue, queueOpts...))
		binding := consumer.Binding
		if binding != nil {
			switch binding.ExchangeType {
			case amqp091.ExchangeDirect:
				opts = append(opts, topology.WithDirectExchange(binding.Exchange))
			case amqp091.ExchangeFanout:
				opts = append(opts, topology.WithFanoutExchange(binding.Exchange))
			case amqp091.ExchangeTopic:
				opts = append(opts, topology.WithTopicExchange(binding.Exchange))
			}

			opts = append(opts, topology.WithBinding(binding.Exchange, consumer.Queue, binding.RoutingKey))
		}
	}

	return topology.New(opts...)
}

type Config struct {
	Url          string
	Publishers   []*publisher.Publisher
	Consumers    []consumer.Consumer
	Declarations topology.Declarations
}

func NewConfig(url string, opts ...ConfigOption) Config {
	cfg := &Config{
		Url: url,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return *cfg
}

func retryPolicyFromConfig(policy RetryPolicy) retry.Policy {
	retries := make([]retry.Retry, 0)
	for _, r := range policy.Retries {
		retries = append(retries, retry.Retry{
			Delay:       time.Duration(r.DelayInMs) * time.Millisecond,
			MaxAttempts: r.MaxAttempts,
		})
	}
	return retry.Policy{
		Retries:          retries,
		FinallyMoveToDlq: policy.FinallyMoveToDlq,
	}
}

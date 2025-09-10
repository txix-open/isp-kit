package grmqx

import (
	"fmt"
	"net/url"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/grmq/retry"
	"github.com/txix-open/grmq/topology"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/rabbitmq_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/rabbitmq/publisher_tracing"
)

type Connection struct {
	Host     string `validate:"required" schema:"Хост"`
	Port     int    `validate:"required" schema:"Порт"`
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
	RoutingKey string `validate:"required" schema:"Ключ маршрутизации,для публикации напрямую в очередь указывается название очереди"`
}

func (p Publisher) DefaultPublisher(restMiddlewares ...publisher.Middleware) *publisher.Publisher {
	middlewares := append(
		[]publisher.Middleware{
			publisher.PersistentMode(),
			PublisherRequestId(),
			PublisherMetrics(rabbitmq_metrics.NewPublisherStorage(metrics.DefaultRegistry)),
			publisher_tracing.NewConfig().Middleware(),
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
	DelayInMs   int `validate:"required" schema:"Задержка в миллисекундах"`
	MaxAttempts int `validate:"required" schema:"Количество попыток,-1 = бесконечно"`
}

type RetryPolicy struct {
	FinallyMoveToDlq bool          `schema:"Отправить в DLQ,отправить сообщение в DLQ в случае последней неудавшеймся попытки обработки"`
	Retries          []RetryConfig `schema:"Настройки"`
}

type Binding struct {
	Exchange     string `validate:"required" schema:"Точка обмена"`
	ExchangeType string `validate:"required,oneof=direct fanout topic" schema:"Тип точки обмена"`
	RoutingKey   string `validate:"required" schema:"Ключ маршрутизации"`
}

type Consumer struct {
	Queue              string         `validate:"required" schema:"Наименование очереди"`
	Dlq                bool           `schema:"Создать очередь DLQ"`
	PrefetchCount      int            `schema:"Количество предзагруженных сообщений,по умолчанию - 1"`
	Concurrency        int            `schema:"Количество обработчиков,по умолчанию - 1, рекомендовано использовать значение = prefetchCount"`
	DisableAutoDeclare bool           `schema:"Отключить автоматическое объявление,по умолчанию  exchange, queue и binding будут созданы автоматически"`
	Binding            *Binding       `schema:"Настройки топологии"`
	RetryPolicy        *RetryPolicy   `schema:"Политика повторной обработки"`
	QueueArgs          map[string]any `schema:"Аргументы очереди"`
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

type BatchConsumer struct {
	Queue              string         `validate:"required" schema:"Наименование очереди"`
	Dlq                bool           `schema:"Создать очередь DLQ"`
	BatchSize          int            `validate:"required" schema:"Количество сообщений в пачке"`
	PurgeIntervalInMs  int            `validate:"required" schema:"Интервал обработки"`
	DisableAutoDeclare bool           `schema:"Отключить автоматическое объявление,по умолчанию  exchange, queue и binding будут созданы автоматически"`
	Binding            *Binding       `schema:"Настройки топологии"`
	RetryPolicy        *RetryPolicy   `schema:"Политика повторной обработки"`
	QueueArgs          map[string]any `schema:"Аргументы очереди"`
}

func (b BatchConsumer) ConsumerConfig() Consumer {
	return Consumer{
		Queue:              b.Queue,
		Dlq:                b.Dlq,
		PrefetchCount:      b.BatchSize,
		Concurrency:        1,
		DisableAutoDeclare: b.DisableAutoDeclare,
		Binding:            b.Binding,
		RetryPolicy:        b.RetryPolicy,
		QueueArgs:          b.QueueArgs,
	}
}

func (b BatchConsumer) DefaultConsumer(handler BatchHandlerAdapter, restMiddlewares ...consumer.Middleware) consumer.Consumer {
	batchHandler := NewBatchHandler(
		handler,
		time.Duration(b.PurgeIntervalInMs)*time.Millisecond,
		b.BatchSize,
	)
	consumer := b.ConsumerConfig().DefaultConsumer(batchHandler, restMiddlewares...)
	consumer.Closer = batchHandler
	return consumer
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
		for k, v := range consumer.QueueArgs {
			switch vv := v.(type) {
			case float64:
				queueOpts = append(queueOpts, topology.WithQueueArg(k, int(vv)))
			default:
				queueOpts = append(queueOpts, topology.WithQueueArg(k, v))
			}
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

func JoinDeclarations(declarations ...topology.Declarations) topology.Declarations {
	result := topology.New()
	for _, declaration := range declarations {
		result.Queues = append(result.Queues, declaration.Queues...)
		result.Exchanges = append(result.Exchanges, declaration.Exchanges...)
		result.Bindings = append(result.Bindings, declaration.Bindings...)
	}
	return result
}

type Config struct {
	Url          string
	Publishers   []*publisher.Publisher
	Consumers    []consumer.Consumer
	Declarations topology.Declarations
	Observer     grmq.Observer
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

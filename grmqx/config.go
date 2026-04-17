package grmqx

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/grmq/retry"
	"github.com/txix-open/grmq/topology"
	"github.com/txix-open/isp-kit/grmqx/batch_handler"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/rabbitmq_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/rabbitmq/publisher_tracing"
)

// Connection represents RabbitMQ connection parameters.
type Connection struct {
	Host     string `validate:"required" schema:"Хост"`
	Port     int    `validate:"required" schema:"Порт"`
	Username string `schema:"Логин"`
	Password string `schema:"Пароль"`
	Vhost    string `schema:"Виртуальный хост"`
}

// Url generates the connection URL for RabbitMQ.
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

// Publisher represents publisher configuration.
type Publisher struct {
	Exchange   string `schema:"Точка обмена"`
	RoutingKey string `validate:"required" schema:"Ключ маршрутизации,для публикации напрямую в очередь указывается название очереди"`
}

// DefaultPublisher creates a publisher with pre-configured middleware and settings:
// - Persistent mode enabled
// - Request ID generation and header injection
// - Metrics and tracing integration
//
// Optional middleware can be provided:
// - PublisherLog: logs published messages
// - PublisherRequestId: generates and injects request IDs (enabled by default)
// - PublisherRetry: adds retry logic on publication errors
// - PublisherMetrics: collects metrics (enabled by default)
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

// RetryConfig represents retry configuration for message processing.
type RetryConfig struct {
	DelayInMs   int `validate:"required" schema:"Задержка в миллисекундах"`
	MaxAttempts int `validate:"required" schema:"Количество попыток,-1 = бесконечно"`
}

// RetryPolicy represents retry policy for message processing.
type RetryPolicy struct {
	FinallyMoveToDlq bool          `schema:"Отправить в DLQ,отправить сообщение в DLQ в случае последней неудавшеймся попытки обработки"`
	Retries          []RetryConfig `schema:"Настройки"`
}

// Binding represents topology binding configuration.
type Binding struct {
	Exchange     string `validate:"required" schema:"Точка обмена"`
	ExchangeType string `validate:"required,oneof=direct fanout topic" schema:"Тип точки обмена"`
	RoutingKey   string `validate:"required" schema:"Ключ маршрутизации"`
}

// Consumer represents consumer configuration.
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

// DefaultConsumer creates a consumer with the specified handler and default settings.
// PrefetchCount and Concurrency default to 1 if not set or less than 1.
// Applies ConsumerRequestId middleware by default.
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

// BatchConsumer represents batch consumer configuration.
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

// ConsumerConfig converts BatchConsumer to a standard Consumer configuration.
// Fixes Concurrency to 1 and inherits all other parameters.
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

// DefaultConsumer creates a batch consumer with batch message processing.
// The handler must implement batch_handler.SyncHandlerAdapter or be convertible
// to batch_handler.SyncHandlerAdapterFunc if it is a function-based handler.
func (b BatchConsumer) DefaultConsumer(handler batch_handler.SyncHandlerAdapter, restMiddlewares ...consumer.Middleware) consumer.Consumer {
	batchHandler := batch_handler.New(
		handler,
		time.Duration(b.PurgeIntervalInMs)*time.Millisecond,
		b.BatchSize,
	)
	consumer := b.ConsumerConfig().DefaultConsumer(batchHandler, restMiddlewares...)
	consumer.Closer = batchHandler
	return consumer
}

// TopologyFromConsumers generates RabbitMQ topology declarations based on consumer configurations.
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

// JoinDeclarations merges multiple topology declarations into one.
func JoinDeclarations(declarations ...topology.Declarations) topology.Declarations {
	result := topology.New()
	for _, declaration := range declarations {
		result.Queues = append(result.Queues, declaration.Queues...)
		result.Exchanges = append(result.Exchanges, declaration.Exchanges...)
		result.Bindings = append(result.Bindings, declaration.Bindings...)
	}
	return result
}

// NewLogObserverFunc is a factory function for creating custom log observers.
type NewLogObserverFunc func(ctx context.Context, logger log.Logger) grmq.Observer

// Config represents the client configuration.
type Config struct {
	Url          string
	Publishers   []*publisher.Publisher
	Consumers    []consumer.Consumer
	Declarations topology.Declarations
	NewObserver  NewLogObserverFunc
}

// NewConfig creates a new configuration with the specified URL and options.
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
	retries := make([]retry.Retry, 0, len(policy.Retries))
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

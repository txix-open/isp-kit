package stompx

import (
	"strconv"

	"github.com/go-stomp/stomp/v3"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/publisher"
)

// Config represents the configuration for a stompx client.
type Config struct {
	// Consumers contains the list of message consumers.
	Consumers []*consumer.Watcher
	// Publishers contains the list of message publishers.
	Publishers []*publisher.Publisher
}

// NewConfig creates a new Config with the provided options.
func NewConfig(opts ...ConfigOption) Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return *cfg
}

// ConsumerConfig holds the configuration for a message consumer.
type ConsumerConfig struct {
	// Address is the broker address (required).
	Address string `validate:"required" schema:"Адрес брокера"`
	// Queue is the queue name (required).
	Queue string `validate:"required" schema:"Очередь"`
	// Concurrency is the number of handlers (default 1).
	Concurrency int `schema:"Кол-во обработчиков,по умолчанию 1"`
	// PrefetchCount is the number of preloaded messages.
	PrefetchCount int `schema:"Кол-во предзагруженных сообщений,по умолчанию не используется"`
	// Username is the username for authentication.
	Username string `schema:"Имя пользователя"`
	// Password is the password for authentication.
	Password string `schema:"Пароль"`
	// ConnHeaders are additional connection headers.
	ConnHeaders map[string]string `schema:"Дополнительные параметры подключения"`
}

// DefaultConsumer creates a consumer configuration with logging, middleware,
// and connection support based on the provided parameters.
func DefaultConsumer(cfg ConsumerConfig, handler consumer.Handler, logger log.Logger, restMiddlewares ...consumer.Middleware) consumer.Config {
	middlewares := []consumer.Middleware{
		ConsumerRequestId(),
	}
	middlewares = append(middlewares, restMiddlewares...)

	concurrency := 1
	if cfg.Concurrency > 0 {
		concurrency = cfg.Concurrency
	}

	connOpts := []consumer.ConnOption{}
	if cfg.Username != "" {
		connOpts = append(connOpts, stomp.ConnOpt.Login(cfg.Username, cfg.Password))
	}
	for key, value := range cfg.ConnHeaders {
		connOpts = append(connOpts, stomp.ConnOpt.Header(key, value))
	}

	subOpts := []consumer.SubscriptionOption{}
	if cfg.PrefetchCount > 0 {
		subOpts = append(subOpts, stomp.SubscribeOpt.Header("activemq.prefetchSize", strconv.Itoa(cfg.PrefetchCount)))
	}

	config := consumer.NewConfig(
		cfg.Address,
		cfg.Queue,
		handler,
		consumer.WithObserver(NewLogObserver(logger)),
		consumer.WithMiddlewares(middlewares...),
		consumer.WithConcurrency(concurrency),
		consumer.WithConnectionOptions(connOpts...),
		consumer.WithSubscriptionOptions(subOpts...),
	)

	return config
}

// PublisherConfig holds the configuration for a message publisher.
type PublisherConfig struct {
	// Address is the broker address (required).
	Address string `validate:"required" schema:"Адрес брокера"`
	// Queue is the queue name (required).
	Queue string `validate:"required" schema:"Очередь"`
	// Username is the username for authentication.
	Username string `schema:"Имя пользователя"`
	// Password is the password for authentication.
	Password string `schema:"Пароль"`
	// ConnHeaders are additional connection headers.
	ConnHeaders map[string]string `schema:"Дополнительные параметры подключения"`
}

// DefaultPublisher creates a message publisher with middleware and connection settings.
func DefaultPublisher(cfg PublisherConfig, restMiddlewares ...publisher.Middleware) *publisher.Publisher {
	middlewares := []publisher.Middleware{
		PublisherPersistent(),
		PublisherRequestId(),
	}
	middlewares = append(middlewares, restMiddlewares...)

	connOpts := []consumer.ConnOption{}
	if cfg.Username != "" {
		connOpts = append(connOpts, stomp.ConnOpt.Login(cfg.Username, cfg.Password))
	}
	for key, value := range cfg.ConnHeaders {
		connOpts = append(connOpts, stomp.ConnOpt.Header(key, value))
	}

	pub := publisher.NewPublisher(
		cfg.Address,
		cfg.Queue,
		publisher.WithMiddlewares(middlewares...),
		publisher.WithConnectionOptions(connOpts...),
	)

	return pub
}

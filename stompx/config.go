package stompx

import (
	"strconv"

	"github.com/go-stomp/stomp/v3"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/publisher"
)

type Config struct {
	Consumers  []*consumer.Watcher
	Publishers []*publisher.Publisher
}

func NewConfig(opts ...ConfigOption) Config {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	return *cfg
}

func (cfg Config) getConsumers() []*consumer.Watcher {
	consumers := make([]*consumer.Watcher, 0)

	for _, c := range cfg.Consumers {
		var newConsumer consumer.Watcher
		newConsumer = *c
		consumers = append(consumers, &newConsumer)
	}

	return consumers
}

type ConsumerConfig struct {
	Address       string            `validate:"required" schema:"Адрес брокера"`
	Queue         string            `validate:"required" schema:"Очередь"`
	Concurrency   int               `schema:"Кол-во обработчиков,по умолчанию 1"`
	PrefetchCount int               `schema:"Кол-во предзагруженных сообщений,по умолчанию не используется"`
	Username      string            `schema:"Имя пользователя"`
	Password      string            `schema:"Пароль"`
	ConnHeaders   map[string]string `schema:"Дополнительные параметры подключения"`
}

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

type PublisherConfig struct {
	Address     string            `validate:"required" schema:"Адрес брокера"`
	Queue       string            `validate:"required" schema:"Очередь"`
	Username    string            `schema:"Имя пользователя"`
	Password    string            `schema:"Пароль"`
	ConnHeaders map[string]string `schema:"Дополнительные параметры подключения"`
}

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

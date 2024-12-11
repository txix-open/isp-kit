package kafkax

import (
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/publisher"
)

const (
	bytesInMb = 1024 * 1024
)

type Config struct {
	Publishers []*publisher.Publisher
	Consumers  []consumer.Consumer
}

func NewConfig(opts ...ConfigOption) Config {
	cfg := &Config{}

	for _, opt := range opts {
		opt(cfg)
	}

	return *cfg
}

type ConfigOption func(c *Config)

func WithConsumers(consumers ...consumer.Consumer) ConfigOption {
	return func(c *Config) {
		c.Consumers = consumers
	}
}

func WithPublishers(publishers ...*publisher.Publisher) ConfigOption {
	return func(c *Config) {
		c.Publishers = publishers
	}
}

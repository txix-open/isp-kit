package kafkax

import (
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/publisher"
)

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

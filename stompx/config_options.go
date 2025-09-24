package stompx

import (
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/publisher"
)

type ConfigOption func(c *Config)

func WithConsumers(consumers ...*consumer.Watcher) ConfigOption {
	return func(c *Config) {
		c.Consumers = consumers
	}
}

func WithPublishers(publishers ...*publisher.Publisher) ConfigOption {
	return func(c *Config) {
		c.Publishers = publishers
	}
}

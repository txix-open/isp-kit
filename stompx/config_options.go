package stompx

import (
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/publisher"
)

// ConfigOption is a function that applies options to a Config.
type ConfigOption func(c *Config)

// WithConsumers adds consumer clients to the client configuration.
func WithConsumers(consumers ...*consumer.Watcher) ConfigOption {
	return func(c *Config) {
		c.Consumers = consumers
	}
}

// WithPublishers adds publisher clients to the client configuration.
func WithPublishers(publishers ...*publisher.Publisher) ConfigOption {
	return func(c *Config) {
		c.Publishers = publishers
	}
}

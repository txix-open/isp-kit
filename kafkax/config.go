// Package kafkax provides a high-level abstraction over Apache Kafka for publishing
// and consuming messages. It builds on top of the franz-go client library and
// includes built-in support for metrics, logging, middlewares, and graceful shutdown.
//
// The package supports both synchronous publishing and concurrent consuming with
// configurable concurrency levels, automatic offset committing, and middleware chains
// for cross-cutting concerns like logging, metrics, and request IDs.
package kafkax

import (
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/publisher"
)

const (
	bytesInMb = 1024 * 1024
)

// Config holds the configuration for Kafka publishers and consumers.
type Config struct {
	Publishers []*publisher.Publisher
	Consumers  []consumer.Consumer
}

// NewConfig creates a new Config instance using the provided options.
func NewConfig(opts ...ConfigOption) Config {
	cfg := &Config{}

	for _, opt := range opts {
		opt(cfg)
	}

	return *cfg
}

// ConfigOption is a function that configures a Config instance.
type ConfigOption func(c *Config)

// WithConsumers sets the consumers for the Config.
func WithConsumers(consumers ...consumer.Consumer) ConfigOption {
	return func(c *Config) {
		c.Consumers = consumers
	}
}

// WithPublishers sets the publishers for the Config.
func WithPublishers(publishers ...*publisher.Publisher) ConfigOption {
	return func(c *Config) {
		c.Publishers = publishers
	}
}

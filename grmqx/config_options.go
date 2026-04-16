package grmqx

import (
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/grmq/topology"
)

// ConfigOption is a function that modifies a Config instance.
type ConfigOption func(c *Config)

// WithConsumers sets the consumers for the configuration.
func WithConsumers(consumers ...consumer.Consumer) ConfigOption {
	return func(c *Config) {
		c.Consumers = consumers
	}
}

// WithPublishers sets the publishers for the configuration.
func WithPublishers(publishers ...*publisher.Publisher) ConfigOption {
	return func(c *Config) {
		c.Publishers = publishers
	}
}

// WithDeclarations sets the topology declarations for the configuration.
func WithDeclarations(declarations topology.Declarations) ConfigOption {
	return func(c *Config) {
		c.Declarations = declarations
	}
}

// WithLogObserver sets a custom log observer factory for the configuration.
func WithLogObserver(newObserverFunc NewLogObserverFunc) ConfigOption {
	return func(c *Config) {
		c.NewObserver = newObserverFunc
	}
}

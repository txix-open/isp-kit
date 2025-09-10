package grmqx

import (
	"github.com/txix-open/grmq"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/grmq/topology"
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

func WithDeclarations(declarations topology.Declarations) ConfigOption {
	return func(c *Config) {
		c.Declarations = declarations
	}
}

func WithLogObserver(observer grmq.Observer) ConfigOption {
	return func(c *Config) {
		c.Observer = observer
	}
}

package grmqx

import (
	"github.com/integration-system/grmq/consumer"
	"github.com/integration-system/grmq/publisher"
	"github.com/integration-system/grmq/topology"
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

package kafkax

type ConfigOption func(c *Config)

func WithConsumers(consumers ...Consumer) ConfigOption {
	return func(c *Config) {
		c.Consumers = consumers
	}
}

func WithPublishers(publishers ...*Publisher) ConfigOption {
	return func(c *Config) {
		c.Publishers = publishers
	}
}

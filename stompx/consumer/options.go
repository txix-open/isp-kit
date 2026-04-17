package consumer

// Option is a function that applies configuration options to a Config.
type Option func(c *Config)

// WithConcurrency sets the number of concurrent workers.
func WithConcurrency(concurrency int) Option {
	return func(c *Config) {
		c.Concurrency = concurrency
	}
}

// WithMiddlewares adds middleware functions to the configuration.
func WithMiddlewares(middlewares ...Middleware) Option {
	return func(c *Config) {
		c.Middlewares = append(c.Middlewares, middlewares...)
	}
}

// WithConnectionOptions adds connection options to the configuration.
func WithConnectionOptions(connOpts ...ConnOption) Option {
	return func(c *Config) {
		c.ConnOpts = append(c.ConnOpts, connOpts...)
	}
}

// WithSubscriptionOptions adds subscription options to the configuration.
func WithSubscriptionOptions(subOpts ...SubscriptionOption) Option {
	return func(c *Config) {
		c.SubscriptionOpts = append(c.SubscriptionOpts, subOpts...)
	}
}

// WithObserver sets the lifecycle observer.
func WithObserver(observer Observer) Option {
	return func(c *Config) {
		c.Observer = observer
	}
}

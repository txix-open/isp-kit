package consumer

type Option func(c *Config)

func WithConcurrency(concurrency int) Option {
	return func(c *Config) {
		c.Concurrency = concurrency
	}
}

func WithMiddlewares(middlewares ...Middleware) Option {
	return func(c *Config) {
		c.Middlewares = append(c.Middlewares, middlewares...)
	}
}

func WithConnectionOptions(connOpts ...ConnOption) Option {
	return func(c *Config) {
		c.ConnOpts = append(c.ConnOpts, connOpts...)
	}
}

func WithSubscriptionOptions(subOpts ...SubscriptionOption) Option {
	return func(c *Config) {
		c.SubscriptionOpts = append(c.SubscriptionOpts, subOpts...)
	}
}

func WithObserver(observer Observer) Option {
	return func(c *Config) {
		c.Observer = observer
	}
}

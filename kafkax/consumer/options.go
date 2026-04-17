package consumer

// Option is a function that configures a Consumer instance.
type Option func(p *Consumer)

// WithMiddlewares configures the consumer with the provided middlewares.
// Middlewares are applied in the order they are provided.
func WithMiddlewares(mws ...Middleware) Option {
	return func(p *Consumer) {
		p.middlewares = append(p.middlewares, mws...)
	}
}

// WithObserver configures the consumer with the provided observer for
// lifecycle event notifications.
func WithObserver(observer Observer) Option {
	return func(c *Consumer) {
		c.observer = observer
	}
}

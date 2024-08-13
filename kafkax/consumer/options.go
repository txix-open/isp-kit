package consumer

type Option func(p *Consumer)

func WithMiddlewares(mws ...Middleware) Option {
	return func(p *Consumer) {
		p.middlewares = append(p.middlewares, mws...)
	}
}

func WithObserver(observer Observer) Option {
	return func(c *Consumer) {
		c.observer = observer
	}
}

package consumer

type Option func(p *Consumer)

func WithMiddlewares(mws ...Middleware) Option {
	return func(p *Consumer) {
		p.Middlewares = append(p.Middlewares, mws...)
	}
}

func WithObserver(observer Observer) Option {
	return func(c *Consumer) {
		c.observer = observer
	}
}

package publisher

type Option func(p *Publisher)

func WithMiddlewares(mws ...Middleware) Option {
	return func(p *Publisher) {
		p.middlewares = append(p.middlewares, mws...)
	}
}

func WithObserver(observer Observer) Option {
	return func(c *Publisher) {
		c.observer = observer
	}
}

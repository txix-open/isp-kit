package publisher

type Option func(p *Publisher)

func WithMiddlewares(mws ...Middleware) Option {
	return func(p *Publisher) {
		p.middlewares = append(p.middlewares, mws...)
	}
}

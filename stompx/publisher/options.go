package publisher

import (
	"gitlab.txix.ru/isp/isp-kit/stompx/consumer"
)

type Option func(p *Publisher)

func WithMiddlewares(mws ...Middleware) Option {
	return func(p *Publisher) {
		p.Middlewares = append(p.Middlewares, mws...)
	}
}

func WithConnectionOptions(connOpts ...consumer.ConnOption) Option {
	return func(p *Publisher) {
		p.ConnOpts = append(p.ConnOpts, connOpts...)
	}
}

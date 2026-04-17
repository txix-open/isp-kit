package publisher

import (
	"github.com/txix-open/isp-kit/stompx/consumer"
)

// Option is a function that applies configuration options to a Publisher.
type Option func(p *Publisher)

// WithMiddlewares adds middleware functions to the publisher.
func WithMiddlewares(mws ...Middleware) Option {
	return func(p *Publisher) {
		p.Middlewares = append(p.Middlewares, mws...)
	}
}

// WithConnectionOptions adds connection options to the publisher.
func WithConnectionOptions(connOpts ...consumer.ConnOption) Option {
	return func(p *Publisher) {
		p.ConnOpts = append(p.ConnOpts, connOpts...)
	}
}

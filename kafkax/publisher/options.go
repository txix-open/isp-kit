package publisher

// Option is a function that configures a Publisher instance.
type Option func(p *Publisher)

// WithMiddlewares configures the publisher with the provided middlewares.
// Middlewares are applied in the order they are provided.
func WithMiddlewares(mws ...Middleware) Option {
	return func(p *Publisher) {
		p.middlewares = append(p.middlewares, mws...)
	}
}

package httpclix

import (
	"github.com/txix-open/isp-kit/http/httpcli"
)

// Default creates a new Client with default middlewares for observability,
// metrics, and request tracing.
//
// Additional options are applied after the default middlewares.
func Default(opts ...httpcli.Option) *httpcli.Client {
	opts = append([]httpcli.Option{
		httpcli.WithMiddlewares(DefaultMiddlewares()...),
	}, opts...)
	return httpcli.New(opts...)
}

// DefaultWithBalancer creates a new ClientBalancer with default middlewares
// and load balancing across the provided hosts.
//
// Additional options are applied after the default middlewares.
func DefaultWithBalancer(initialHosts []string, opts ...Option) *ClientBalancer {
	opts = append([]Option{
		WithClientOptions(
			httpcli.WithMiddlewares(DefaultMiddlewares()...),
		),
	}, opts...)
	return NewClientBalancer(initialHosts, opts...)
}

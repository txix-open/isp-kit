package httpclix

import (
	"github.com/txix-open/isp-kit/http/httpcli"
)

func Default(opts ...httpcli.Option) *httpcli.Client {
	opts = append([]httpcli.Option{
		httpcli.WithMiddlewares(DefaultMiddlewares()...),
	}, opts...)
	return httpcli.New(opts...)
}

func DefaultWithBalancer(initialHosts []string, opts ...Option) *ClientBalancer {
	opts = append([]Option{
		WithClientOptions(
			httpcli.WithMiddlewares(DefaultMiddlewares()...),
		),
	}, opts...)
	return NewClientBalancer(initialHosts, opts...)
}

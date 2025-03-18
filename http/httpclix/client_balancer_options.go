package httpclix

import (
	"net/http"

	"github.com/txix-open/isp-kit/http/httpcli"
)

type clientBalancerOptions struct {
	cli        *http.Client
	clientOpts []httpcli.Option
	schema     string
}

type Option func(c *clientBalancerOptions)

// WithClientOptions appends passed opts to clientOpts
func WithClientOptions(opts ...httpcli.Option) Option {
	return func(c *clientBalancerOptions) {
		c.clientOpts = append(c.clientOpts, opts...)
	}
}

func WithHttpsSchema() Option {
	return func(c *clientBalancerOptions) {
		c.schema = httpsSchema
	}
}

func WithClient(cli *http.Client) Option {
	return func(c *clientBalancerOptions) {
		c.cli = cli
	}
}

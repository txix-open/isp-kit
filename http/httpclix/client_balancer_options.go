package httpclix

import (
	"net/http"

	"github.com/txix-open/isp-kit/http/httpcli"
)

// clientBalancerOptions holds configuration options for ClientBalancer.
type clientBalancerOptions struct {
	cli        *http.Client
	clientOpts []httpcli.Option
	schema     string
}

// Option is a function that configures a ClientBalancer.
type Option func(c *clientBalancerOptions)

// WithClientOptions appends the given client options to the ClientBalancer.
func WithClientOptions(opts ...httpcli.Option) Option {
	return func(c *clientBalancerOptions) {
		c.clientOpts = append(c.clientOpts, opts...)
	}
}

// WithHttpsSchema configures the ClientBalancer to use https:// as the URL schema.
func WithHttpsSchema() Option {
	return func(c *clientBalancerOptions) {
		c.schema = httpsSchema
	}
}

// WithClient sets a custom http.Client for the ClientBalancer.
func WithClient(cli *http.Client) Option {
	return func(c *clientBalancerOptions) {
		c.cli = cli
	}
}

package httpclix

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/lb"
)

const (
	httpsSchema = "https://"
	httpSchema  = "http://"
)

// ClientBalancer is an HTTP client that distributes requests across multiple
// hosts using round-robin load balancing.
//
// It embeds httpcli.Client and extends it with host management functionality.
type ClientBalancer struct {
	*httpcli.Client

	hostManager *lb.RoundRobin
	schema      string
}

// NewClientBalancer creates a new ClientBalancer with the given initial hosts
// and options.
//
// Hosts without a schema prefix automatically get http:// prepended.
func NewClientBalancer(initialHosts []string, opts ...Option) *ClientBalancer {
	options := &clientBalancerOptions{
		cli:        nil,
		clientOpts: nil,
		schema:     httpSchema,
	}
	for _, opt := range opts {
		opt(options)
	}
	initialHosts = addSchemaToHosts(options.schema, initialHosts)

	return &ClientBalancer{
		Client:      httpClient(options.cli, options.clientOpts...),
		hostManager: lb.NewRoundRobin(initialHosts),
		schema:      options.schema,
	}
}

// Post creates a new POST request builder for the given path.
// The path is appended to the currently selected host.
func (c *ClientBalancer) Post(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodPost, method, c.GlobalRequestConfig(), c.Execute)
}

// Get creates a new GET request builder for the given path.
// The path is appended to the currently selected host.
func (c *ClientBalancer) Get(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodGet, method, c.GlobalRequestConfig(), c.Execute)
}

// Put creates a new PUT request builder for the given path.
// The path is appended to the currently selected host.
func (c *ClientBalancer) Put(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodPut, method, c.GlobalRequestConfig(), c.Execute)
}

// Delete creates a new DELETE request builder for the given path.
// The path is appended to the currently selected host.
func (c *ClientBalancer) Delete(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodDelete, method, c.GlobalRequestConfig(), c.Execute)
}

// Patch creates a new PATCH request builder for the given path.
// The path is appended to the currently selected host.
func (c *ClientBalancer) Patch(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodPatch, method, c.GlobalRequestConfig(), c.Execute)
}

// Execute sends an HTTP request using the provided builder and the load balancer.
//
// If GlobalRequestConfig.BaseUrl is set, the request is sent to that URL.
// Otherwise, the next host from the round-robin balancer is selected and the
// request path is appended to it.
func (c *ClientBalancer) Execute(ctx context.Context, builder *httpcli.RequestBuilder) (*httpcli.Response, error) {
	if c.GlobalRequestConfig().BaseUrl != "" {
		return c.Client.Execute(ctx, builder)
	}

	host, err := c.hostManager.Next()
	if err != nil {
		return nil, errors.WithMessage(err, "host manager next")
	}

	return c.Client.Execute(ctx, builder.BaseUrl(host))
}

// Upgrade replaces the current set of hosts with a new set.
//
// Hosts without a schema prefix automatically get the configured schema prepended.
func (c *ClientBalancer) Upgrade(hosts []string) {
	hosts = addSchemaToHosts(c.schema, hosts)
	c.hostManager.Upgrade(hosts)
}

// addSchemaToHosts prepends the schema to hosts that don't have one.
func addSchemaToHosts(schema string, hosts []string) []string {
	for i, host := range hosts {
		shouldAddSchema := !strings.HasPrefix(host, httpSchema) && !strings.HasPrefix(host, httpsSchema)
		if shouldAddSchema {
			hosts[i] = schema + host
		}
	}
	return hosts
}

// httpClient creates an httpcli.Client, using the provided client if non-nil,
// otherwise creating a new one with the given options.
func httpClient(cli *http.Client, opts ...httpcli.Option) *httpcli.Client {
	if cli == nil {
		return httpcli.New(opts...)
	}
	return httpcli.NewWithClient(cli, opts...)
}

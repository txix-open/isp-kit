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

type ClientBalancer struct {
	*httpcli.Client
	hostManager *lb.RoundRobin
	schema      string
}

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

func (c *ClientBalancer) Post(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodPost, method, c.GlobalRequestConfig(), c.Execute)
}

func (c *ClientBalancer) Get(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodGet, method, c.GlobalRequestConfig(), c.Execute)
}

func (c *ClientBalancer) Put(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodPut, method, c.GlobalRequestConfig(), c.Execute)
}

func (c *ClientBalancer) Delete(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodDelete, method, c.GlobalRequestConfig(), c.Execute)
}

func (c *ClientBalancer) Patch(method string) *httpcli.RequestBuilder {
	return httpcli.NewRequestBuilder(http.MethodPatch, method, c.GlobalRequestConfig(), c.Execute)
}

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

func (c *ClientBalancer) Upgrade(hosts []string) {
	hosts = addSchemaToHosts(c.schema, hosts)
	c.hostManager.Upgrade(hosts)
}

func addSchemaToHosts(schema string, hosts []string) []string {
	for i, host := range hosts {
		shouldAddSchema := !strings.HasPrefix(host, httpSchema) && !strings.HasPrefix(host, httpsSchema)
		if shouldAddSchema {
			hosts[i] = schema + host
		}
	}
	return hosts
}

func httpClient(cli *http.Client, opts ...httpcli.Option) *httpcli.Client {
	if cli == nil {
		return httpcli.New(opts...)
	}
	return httpcli.NewWithClient(cli, opts...)
}

// Package client provides a gRPC client implementation with built-in load balancing,
// middleware support, and observability features.
//
// The client supports dynamic host updates for service discovery and integrates with
// the ISP kit framework's logging, metrics, and tracing systems.
package client

import (
	"context"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/grpc/client/request"
	"github.com/txix-open/isp-kit/grpc/isp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/manual"
)

const (
	// resolverScheme defines the custom resolver scheme for host discovery.
	resolverScheme = "isp"
	// resolverUrl is the target URL for the manual resolver.
	resolverUrl = resolverScheme + ":///"
)

// Client is a gRPC client with support for middleware, load balancing, and dynamic host updates.
// It provides a fluent API for building requests and automatically handles connection management.
// Client is safe for concurrent use by multiple goroutines.
type Client struct {
	middlewares []request.Middleware
	dialOptions []grpc.DialOption

	roundTripper  request.RoundTripper
	hostsResolver *manual.Resolver
	grpcCli       *grpc.ClientConn
	backendCli    isp.BackendServiceClient

	currentHosts atomic.Value
}

// New creates a new Client with the specified initial hosts and options.
// Returns an error if the gRPC client cannot be initialized.
// The client automatically connects to the provided hosts and applies all configured middleware.
func New(initialHosts []string, opts ...Option) (*Client, error) {
	cli := &Client{}
	for _, opt := range opts {
		opt(cli)
	}

	hostsResolver := manual.NewBuilderWithScheme(resolverScheme)
	hostsResolver.InitialState(resolver.State{
		Addresses: toAddresses(initialHosts),
	})
	dialOptions := cli.dialOptions
	dialOptions = append(
		dialOptions,
		grpc.WithResolvers(hostsResolver),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
	)

	grpcCli, err := grpc.NewClient(resolverUrl, dialOptions...)
	if err != nil {
		return nil, errors.WithMessage(err, "new grpc client")
	}
	grpcCli.Connect()
	backendCli := isp.NewBackendServiceClient(grpcCli)

	cli.currentHosts = atomic.Value{}
	cli.currentHosts.Store(initialHosts)
	cli.hostsResolver = hostsResolver
	cli.backendCli = backendCli
	cli.grpcCli = grpcCli

	roundTripper := cli.do
	for i := len(cli.middlewares) - 1; i >= 0; i-- {
		roundTripper = cli.middlewares[i](roundTripper)
	}
	cli.roundTripper = roundTripper

	return cli, nil
}

// Invoke creates a request builder for the specified endpoint.
// The builder provides a fluent API for configuring and executing the request.
func (cli *Client) Invoke(endpoint string) *request.Builder {
	return request.NewBuilder(cli.roundTripper, endpoint)
}

// Upgrade updates the list of backend hosts for load balancing.
// This enables dynamic service discovery without restarting the client.
// Thread-safe for concurrent use.
func (cli *Client) Upgrade(hosts []string) {
	cli.currentHosts.Store(hosts)
	cli.hostsResolver.UpdateState(resolver.State{
		Addresses: toAddresses(hosts),
	})
}

// Close closes the gRPC client connection.
// Should be called when the client is no longer needed.
func (cli *Client) Close() error {
	return cli.grpcCli.Close()
}

// BackendClient returns the underlying gRPC BackendServiceClient.
// Useful for direct gRPC calls bypassing the client abstraction.
func (cli *Client) BackendClient() isp.BackendServiceClient {
	return cli.backendCli
}

// do executes the actual gRPC request through the middleware chain.
// Returns an error if the client is not properly initialized or the request fails.
func (cli *Client) do(ctx context.Context, _ *request.Builder, message *isp.Message) (*isp.Message, error) {
	currentHosts := cli.currentHosts.Load().([]string) // nolint:forcetypeassert
	if len(currentHosts) == 0 {
		return nil, errors.New("grpc client: client is not initialized properly: empty hosts array")
	}
	return cli.backendCli.Request(ctx, message)
}

// toAddresses converts a slice of host strings to gRPC resolver addresses.
func toAddresses(hosts []string) []resolver.Address {
	addresses := make([]resolver.Address, len(hosts))
	for i, host := range hosts {
		addresses[i] = resolver.Address{
			Addr: host,
		}
	}
	return addresses
}

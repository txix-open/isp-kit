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
	resolverScheme = "isp"
	resolverUrl    = resolverScheme + ":///"
)

type Client struct {
	middlewares []request.Middleware
	dialOptions []grpc.DialOption

	roundTripper  request.RoundTripper
	hostsResolver *manual.Resolver
	grpcCli       *grpc.ClientConn
	backendCli    isp.BackendServiceClient

	currentHosts atomic.Value
}

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

func (cli *Client) Invoke(endpoint string) *request.Builder {
	return request.NewBuilder(cli.roundTripper, endpoint)
}

func (cli *Client) Upgrade(hosts []string) {
	cli.currentHosts.Store(hosts)
	cli.hostsResolver.UpdateState(resolver.State{
		Addresses: toAddresses(hosts),
	})
}

func (cli *Client) Close() error {
	return cli.grpcCli.Close()
}

func (cli *Client) BackendClient() isp.BackendServiceClient {
	return cli.backendCli
}

func (cli *Client) do(ctx context.Context, _ *request.Builder, message *isp.Message) (*isp.Message, error) {
	currentHosts := cli.currentHosts.Load().([]string)
	if len(currentHosts) == 0 {
		return nil, errors.New("grpc client: client is not initialized properly: empty hosts array")
	}
	return cli.backendCli.Request(ctx, message)
}

func toAddresses(hosts []string) []resolver.Address {
	addresses := make([]resolver.Address, len(hosts))
	for i, host := range hosts {
		addresses[i] = resolver.Address{
			Addr: host,
		}
	}
	return addresses
}

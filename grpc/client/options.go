package client

import (
	"github.com/txix-open/isp-kit/grpc/client/request"
	"google.golang.org/grpc"
)

type Option func(cli *Client)

func WithMiddlewares(middlewares ...request.Middleware) Option {
	return func(cli *Client) {
		cli.middlewares = append(cli.middlewares, middlewares...)
	}
}

func WithDialOptions(dialOptions ...grpc.DialOption) Option {
	return func(cli *Client) {
		cli.dialOptions = dialOptions
	}
}

package client

import (
	"google.golang.org/grpc"
)

type Option func(cli *Client)

func WithMiddlewares(middlewares ...Middleware) Option {
	return func(cli *Client) {
		cli.middlewares = middlewares
	}
}

func WithDialOptions(dialOptions ...grpc.DialOption) Option {
	return func(cli *Client) {
		cli.dialOptions = dialOptions
	}
}

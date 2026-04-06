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

type LogOption func(cfg *logConfig)

func WithLogBody(logBody bool) LogOption {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logBody
		cfg.logRequestBody = logBody
	}
}

func WithLogResponseBody(logResponseBody bool) LogOption {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logResponseBody
	}
}

func WithLogRequestBody(logRequestBody bool) LogOption {
	return func(cfg *logConfig) {
		cfg.logRequestBody = logRequestBody
	}
}

func WithCombinedLog(enable bool) LogOption {
	return func(cfg *logConfig) {
		cfg.combinedLog = enable
	}
}

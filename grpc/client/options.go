package client

import (
	"github.com/txix-open/isp-kit/grpc/client/request"
	"google.golang.org/grpc"
)

// Option configures a Client using the functional options pattern.
type Option func(cli *Client)

// WithMiddlewares adds one or more request middleware to the client.
// Middleware are executed in the order they are provided.
func WithMiddlewares(middlewares ...request.Middleware) Option {
	return func(cli *Client) {
		cli.middlewares = append(cli.middlewares, middlewares...)
	}
}

// WithDialOptions sets custom gRPC dial options for the client.
// Note: This replaces any previously set dial options.
func WithDialOptions(dialOptions ...grpc.DialOption) Option {
	return func(cli *Client) {
		cli.dialOptions = dialOptions
	}
}

// LogOption configures logging behavior for request middleware.
type LogOption func(cfg *logConfig)

// WithLogBody enables logging of both request and response bodies.
func WithLogBody(logBody bool) LogOption {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logBody
		cfg.logRequestBody = logBody
	}
}

// WithLogResponseBody enables or disables logging of response bodies.
func WithLogResponseBody(logResponseBody bool) LogOption {
	return func(cfg *logConfig) {
		cfg.logResponseBody = logResponseBody
	}
}

// WithLogRequestBody enables or disables logging of request bodies.
func WithLogRequestBody(logRequestBody bool) LogOption {
	return func(cfg *logConfig) {
		cfg.logRequestBody = logRequestBody
	}
}

// WithCombinedLog enables a single combined log entry for request and response.
// When disabled, requests and responses are logged separately.
func WithCombinedLog(enable bool) LogOption {
	return func(cfg *logConfig) {
		cfg.combinedLog = enable
	}
}

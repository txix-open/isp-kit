package client

import (
	ispgrpc "github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/client/request"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/grpc_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/grpc/client_tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Default creates a Client with pre-configured middleware for observability.
// Includes request ID propagation, metrics collection, and distributed tracing.
// Uses insecure transport by default (suitable for development and testing).
// Accepts additional middleware to be appended after the default ones.
// Returns an error if the client cannot be initialized.
func Default(restMiddlewares ...request.Middleware) (*Client, error) {
	middlewares := append(
		[]request.Middleware{
			RequestId(),
			Metrics(grpc_metrics.NewClientStorage(metrics.DefaultRegistry)),
			client_tracing.NewConfig().Middleware(),
		},
		restMiddlewares...,
	)
	return New(
		nil,
		WithDialOptions(
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultCallOptions(
				grpc.WaitForReady(true),
				grpc.MaxCallSendMsgSize(ispgrpc.DefaultMaxSizeByte),
				grpc.MaxCallRecvMsgSize(ispgrpc.DefaultMaxSizeByte),
			),
		),
		WithMiddlewares(middlewares...),
	)
}

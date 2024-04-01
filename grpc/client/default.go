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

package client

import (
	ispgrpc "github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/metrics/grpc_metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Default(restMiddlewares ...Middleware) (*Client, error) {
	middlewares := append(
		[]Middleware{
			RequestId(),
			Metrics(grpc_metrics.NewClientStorage(metrics.DefaultRegistry)),
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

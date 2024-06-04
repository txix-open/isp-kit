package client

import (
	ispgrpc "gitlab.txix.ru/isp/isp-kit/grpc"
	"gitlab.txix.ru/isp/isp-kit/grpc/client/request"
	"gitlab.txix.ru/isp/isp-kit/metrics"
	"gitlab.txix.ru/isp/isp-kit/metrics/grpc_metrics"
	"gitlab.txix.ru/isp/isp-kit/observability/tracing/grpc/client_tracing"
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

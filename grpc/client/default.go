package client

import (
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/metrics/grpc_metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultMaxSizeByte = 64 * 1024 * 1024
)

func Default(restMiddlewares ...Middleware) (*Client, error) {
	middlewares := append(
		[]Middleware{
			RequestId(),
			DefaultTimeout(15 * time.Second),
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
				grpc.MaxCallSendMsgSize(defaultMaxSizeByte),
				grpc.MaxCallRecvMsgSize(defaultMaxSizeByte),
			),
		),
		WithMiddlewares(middlewares...),
	)
}

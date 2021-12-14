package client

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultMaxSizeByte = 10 * 1024 * 1024
)

func Default() (*Client, error) {
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
		WithMiddlewares(
			RequestId(),
			DefaultTimeout(15*time.Second),
		),
	)
}

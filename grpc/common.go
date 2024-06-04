package grpc

import (
	"context"

	"gitlab.txix.ru/isp/isp-kit/grpc/isp"
)

type HandlerFunc func(ctx context.Context, message *isp.Message) (*isp.Message, error)

type Middleware func(next HandlerFunc) HandlerFunc

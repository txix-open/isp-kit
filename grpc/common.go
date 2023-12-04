package grpc

import (
	"context"

	"github.com/integration-system/isp-kit/grpc/isp"
)

type HandlerFunc func(ctx context.Context, message *isp.Message) (*isp.Message, error)

type Middleware func(next HandlerFunc) HandlerFunc

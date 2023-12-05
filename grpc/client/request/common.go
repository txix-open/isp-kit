package request

import (
	"context"

	"github.com/integration-system/isp-kit/grpc/isp"
)

type RoundTripper func(ctx context.Context, builder *Builder, message *isp.Message) (*isp.Message, error)

type Middleware func(next RoundTripper) RoundTripper

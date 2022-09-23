package client

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/requestid"
	"google.golang.org/grpc/metadata"
)

func RequestId() Middleware {
	return func(next RoundTripper) RoundTripper {
		return func(ctx context.Context, builder *RequestBuilder, message *isp.Message) (*isp.Message, error) {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}

			ctx = metadata.AppendToOutgoingContext(ctx, grpc.RequestIdHeader, requestId)
			return next(ctx, message)
		}
	}
}

func DefaultTimeout(timeout time.Duration) Middleware {
	return func(next RoundTripper) RoundTripper {
		return func(ctx context.Context, builder *RequestBuilder, message *isp.Message) (*isp.Message, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			return next(ctx, message)
		}
	}
}

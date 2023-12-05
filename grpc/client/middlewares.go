package client

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/grpc"
	"github.com/integration-system/isp-kit/grpc/client/request"
	"github.com/integration-system/isp-kit/grpc/isp"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/requestid"
	"google.golang.org/grpc/metadata"
)

func RequestId() request.Middleware {
	return func(next request.RoundTripper) request.RoundTripper {
		return func(ctx context.Context, builder *request.Builder, message *isp.Message) (*isp.Message, error) {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}

			ctx = metadata.AppendToOutgoingContext(ctx, grpc.RequestIdHeader, requestId)
			return next(ctx, builder, message)
		}
	}
}

func Log(logger log.Logger) request.Middleware {
	return func(next request.RoundTripper) request.RoundTripper {
		return func(ctx context.Context, builder *request.Builder, message *isp.Message) (*isp.Message, error) {
			logger.Debug(
				ctx,
				"grpc client: request",
				log.String("requestEndpoint", builder.Endpoint),
				log.ByteString("requestBody", message.GetBytesBody()),
			)

			resp, err := next(ctx, builder, message)
			if err != nil {
				logger.Debug(ctx, "grpc client: response", log.Any("error", err))
				return resp, err
			}

			logger.Debug(
				ctx,
				"grpc client: response",
				log.ByteString("responseBody", resp.GetBytesBody()),
			)

			return resp, err
		}
	}
}

type MetricStorage interface {
	ObserveDuration(endpoint string, duration time.Duration)
}

func Metrics(storage MetricStorage) request.Middleware {
	return func(next request.RoundTripper) request.RoundTripper {
		return func(ctx context.Context, builder *request.Builder, message *isp.Message) (*isp.Message, error) {
			start := time.Now()
			response, err := next(ctx, builder, message)
			storage.ObserveDuration(builder.Endpoint, time.Since(start))
			return response, err
		}
	}
}

package client

import (
	"context"
	"time"

	"github.com/txix-open/isp-kit/grpc/client/request"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"google.golang.org/grpc/metadata"
)

func RequestId() request.Middleware {
	return func(next request.RoundTripper) request.RoundTripper {
		return func(ctx context.Context, builder *request.Builder, message *isp.Message) (*isp.Message, error) {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}

			ctx = metadata.AppendToOutgoingContext(ctx, requestid.Header, requestId)
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

			now := time.Now()
			resp, err := next(ctx, builder, message)
			if err != nil {
				logger.Debug(
					ctx,
					"grpc client: response with error",
					log.Any("error", err),
					log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
				)
				return resp, err
			}

			logger.Debug(
				ctx,
				"grpc client: response",
				log.ByteString("responseBody", resp.GetBytesBody()),
				log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
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

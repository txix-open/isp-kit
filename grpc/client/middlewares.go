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

func Log(logger log.Logger, logBody bool) request.Middleware {
	return func(next request.RoundTripper) request.RoundTripper {
		return func(ctx context.Context, builder *request.Builder, message *isp.Message) (*isp.Message, error) {
			requestFields := []log.Field{
				log.String("endpoint", builder.Endpoint),
			}
			if logBody {
				requestFields = append(requestFields, log.ByteString("requestBody", message.GetBytesBody()))
			}
			logger.Debug(
				ctx,
				"grpc client: request",
				requestFields...,
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

			responseFields := []log.Field{
				log.String("endpoint", builder.Endpoint),
				log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
			}
			if logBody {
				responseFields = append(responseFields, log.ByteString("responseBody", resp.GetBytesBody()))
			}
			logger.Debug(
				ctx,
				"grpc client: response",
				responseFields...,
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

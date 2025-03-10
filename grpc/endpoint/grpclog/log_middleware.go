package grpclog

import (
	"context"
	"time"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/log"
)

type logConfig struct {
	logRequestBody  bool
	logResponseBody bool
}

func Log(logger log.Logger, opts ...Option) grpc.Middleware {
	cfg := &logConfig{
		logRequestBody:  false,
		logResponseBody: false,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	return middleware(logger, cfg)
}

func middleware(logger log.Logger, cfg *logConfig) grpc.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			requestFields := []log.Field{}
			if cfg.logRequestBody {
				requestFields = append(requestFields, log.ByteString("requestBody", message.GetBytesBody()))
			}
			logger.Debug(ctx, "grpc handler: request", requestFields...)

			now := time.Now()
			response, err := next(ctx, message)
			if err != nil {
				return response, err
			}

			responseFields := []log.Field{
				log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
			}
			if cfg.logResponseBody {
				responseFields = append(responseFields, log.ByteString("responseBody", response.GetBytesBody()))
			}
			logger.Debug(ctx,
				"grpc handler: response",
				responseFields...,
			)
			return response, err
		}
	}
}

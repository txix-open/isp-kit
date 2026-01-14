package grpclog

import (
	"context"
	"time"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/log"
	"google.golang.org/grpc/metadata"
)

type logConfig struct {
	logRequestBody  bool
	logResponseBody bool
	combinedLog     bool
}

func Log(logger log.Logger, logBody bool) grpc.Middleware {
	cfg := &logConfig{
		logRequestBody:  logBody,
		logResponseBody: logBody,
	}
	return middleware(logger, cfg)
}

func CombinedLog(logger log.Logger, logBody bool) grpc.Middleware {
	cfg := &logConfig{
		logRequestBody:  logBody,
		logResponseBody: logBody,
	}
	return combinedLogMiddleware(logger, cfg)
}

func LogWithOptions(logger log.Logger, opts ...Option) grpc.Middleware {
	cfg := &logConfig{
		logRequestBody:  false,
		logResponseBody: false,
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.combinedLog {
		return combinedLogMiddleware(logger, cfg)
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

func combinedLogMiddleware(logger log.Logger, cfg *logConfig) grpc.Middleware {
	return func(next grpc.HandlerFunc) grpc.HandlerFunc {
		return func(ctx context.Context, message *isp.Message) (*isp.Message, error) {
			logFields := []log.Field{}
			if cfg.logRequestBody {
				logFields = append(logFields, log.ByteString("requestBody", message.GetBytesBody()))
			}
			logFields = append(logFields, applicationLogFields(ctx)...)

			now := time.Now()
			response, err := next(ctx, message)
			if err != nil {
				return response, err
			}

			logFields = append(logFields,
				log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
			)
			if cfg.logResponseBody {
				logFields = append(logFields, log.ByteString("responseBody", response.GetBytesBody()))
			}
			logger.Debug(ctx,
				"grpc handler: log",
				logFields...,
			)
			return response, err
		}
	}
}

func applicationLogFields(ctx context.Context) []log.Field {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	authData := grpc.AuthData(md)

	logFields := make([]log.Field, 0)
	appName, err := authData.ApplicationName()
	if err == nil {
		logFields = append(logFields, log.String("applicationName", appName))
	}

	appId, err := authData.ApplicationId()
	if err == nil {
		logFields = append(logFields, log.Int("applicationId", appId))
	}

	return logFields
}

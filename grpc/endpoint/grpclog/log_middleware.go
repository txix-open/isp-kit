// Package grpclog provides logging middleware for gRPC server handlers.
// It supports request/response body logging with configurable granularity
// and automatic inclusion of application context from metadata.
package grpclog

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/log"
	"google.golang.org/grpc/metadata"
)

// logConfig holds logging configuration for server middleware.
type logConfig struct {
	logRequestBody  bool
	logResponseBody bool
	combinedLog     bool
}

// Log creates a middleware that logs gRPC server requests and responses separately.
// When logBody is true, request and response bodies are included in the logs.
// Logs at Debug level for requests and responses.
func Log(logger log.Logger, logBody bool) grpc.Middleware {
	cfg := &logConfig{
		logRequestBody:  logBody,
		logResponseBody: logBody,
	}
	return middleware(logger, cfg)
}

// CombinedLog creates a middleware that logs gRPC server requests and responses in a single entry.
// When logBody is true, request and response bodies are included in the logs.
// Includes application context (name and ID) from metadata.
func CombinedLog(logger log.Logger, logBody bool) grpc.Middleware {
	cfg := &logConfig{
		logRequestBody:  logBody,
		logResponseBody: logBody,
	}
	return combinedLogMiddleware(logger, cfg)
}

// LogWithOptions creates a middleware that logs gRPC server requests and responses with custom options.
// Provides fine-grained control over what is logged (request body, response body, combined logs).
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

// middleware implements request/response logging with separate log entries.
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

// combinedLogMiddleware implements logging with a single combined log entry.
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

// applicationLogFields extracts application context from gRPC metadata.
// Returns application name (decoded from base64) and application ID if available.
func applicationLogFields(ctx context.Context) []log.Field {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	authData := grpc.AuthData(md)

	logFields := make([]log.Field, 0)
	appName, err := authData.ApplicationName()
	if err == nil {
		decodedName, err := base64.StdEncoding.DecodeString(appName)
		if err == nil {
			logFields = append(logFields, log.String("applicationName", string(decodedName)))
		}
	}

	appId, err := authData.ApplicationId()
	if err == nil {
		logFields = append(logFields, log.Int("applicationId", appId))
	}

	return logFields
}

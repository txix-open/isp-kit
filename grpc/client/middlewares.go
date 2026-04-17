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

// logConfig holds logging configuration for client middleware.
type logConfig struct {
	logRequestBody  bool
	logResponseBody bool
	combinedLog     bool
}

// RequestId is a middleware that propagates request IDs across service boundaries.
// If no request ID is present in the context, it generates a new one.
// The request ID is added to outgoing metadata for tracing purposes.
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

// Log creates a middleware that logs gRPC client requests and responses.
// When logBody is true, request and response bodies are included in the logs.
// Logs at Debug level for requests and responses.
func Log(logger log.Logger, logBody bool) request.Middleware {
	cfg := &logConfig{
		logRequestBody:  logBody,
		logResponseBody: logBody,
	}
	return logMiddleware(logger, cfg)
}

// LogWithOptions creates a middleware that logs gRPC client requests and responses with custom options.
// Provides fine-grained control over what is logged (request body, response body, combined logs).
func LogWithOptions(logger log.Logger, opts ...LogOption) request.Middleware {
	cfg := &logConfig{
		logRequestBody:  false,
		logResponseBody: false,
		combinedLog:     false,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.combinedLog {
		return logCombinedMiddleware(logger, cfg)
	}
	return logMiddleware(logger, cfg)
}

// logMiddleware implements request/response logging with separate log entries.
func logMiddleware(logger log.Logger, cfg *logConfig) request.Middleware {
	return func(next request.RoundTripper) request.RoundTripper {
		return func(ctx context.Context, builder *request.Builder, message *isp.Message) (*isp.Message, error) {
			requestFields := []log.Field{
				log.String("endpoint", builder.Endpoint),
			}
			if cfg.logRequestBody {
				requestFields = append(requestFields, log.ByteString("requestBody", message.GetBytesBody()))
			}
			logger.Debug(ctx, "grpc client: request", requestFields...)

			var responseFields []log.Field

			now := time.Now()
			resp, err := next(ctx, builder, message)
			if err != nil {
				responseFields = append(responseFields,
					log.Any("error", err),
					log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
				)
				logger.Debug(ctx, "grpc client: response with error", responseFields...)

				return resp, err
			}

			responseFields = append(responseFields,
				log.String("endpoint", builder.Endpoint),
				log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
			)
			if cfg.logResponseBody {
				responseFields = append(responseFields, log.ByteString("responseBody", resp.GetBytesBody()))
			}

			logger.Debug(ctx, "grpc client: response", responseFields...)

			return resp, err
		}
	}
}

// logCombinedMiddleware implements logging with a single combined log entry.
func logCombinedMiddleware(logger log.Logger, cfg *logConfig) request.Middleware {
	return func(next request.RoundTripper) request.RoundTripper {
		return func(ctx context.Context, builder *request.Builder, message *isp.Message) (*isp.Message, error) {
			logFields := []log.Field{
				log.String("endpoint", builder.Endpoint),
			}
			if cfg.logRequestBody {
				logFields = append(logFields, log.ByteString("requestBody", message.GetBytesBody()))
			}

			now := time.Now()
			resp, err := next(ctx, builder, message)

			logFields = append(logFields,
				log.String("endpoint", builder.Endpoint),
				log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
			)

			if err != nil {
				logFields = append(logFields, log.Any("error", err))
				logger.Debug(ctx, "grpc client: log with error", logFields...)
				return resp, err
			}

			if cfg.logResponseBody {
				logFields = append(logFields, log.ByteString("responseBody", resp.GetBytesBody()))
			}
			logger.Debug(ctx, "grpc client: log", logFields...)

			return resp, err
		}
	}
}

// MetricStorage defines the interface for metric storage implementations.
// Used by the Metrics middleware to collect timing information.
type MetricStorage interface {
	// ObserveDuration records the duration of a request for the given endpoint.
	ObserveDuration(endpoint string, duration time.Duration)
}

// Metrics creates a middleware that collects timing metrics for gRPC client requests.
// The storage implementation receives the endpoint name and request duration.
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

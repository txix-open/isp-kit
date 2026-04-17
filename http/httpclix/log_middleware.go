package httpclix

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
)

// logConfig holds configuration for the logging middleware.
type logConfig struct {
	LogRequestBody  bool
	LogResponseBody bool

	LogDumpRequest  bool
	LogDumpResponse bool

	LogHeadersRequest  bool
	LogHeadersResponse bool

	CombinedLog bool
}

// LogOption is a function that configures log behavior.
type LogOption func(*logConfig)

// LogDump enables raw HTTP request/response dumping.
// When enabled, the full HTTP request and/or response are logged including headers and body.
func LogDump(dumpRequest bool, dumpResponse bool) LogOption {
	return func(cfg *logConfig) {
		cfg.LogDumpRequest = dumpRequest
		cfg.LogDumpResponse = dumpResponse
	}
}

// LogHeaders enables logging of HTTP request and response headers.
func LogHeaders(requestHeaders bool, responseHeaders bool) LogOption {
	return func(cfg *logConfig) {
		cfg.LogHeadersRequest = requestHeaders
		cfg.LogHeadersResponse = responseHeaders
	}
}

// LogCombined enables combined logging mode where all fields are logged in a single
// log entry instead of separate request and response entries.
func LogCombined(combinedLog bool) LogOption {
	return func(cfg *logConfig) {
		cfg.CombinedLog = combinedLog
	}
}

// logConfigContextKey is the context key for per-request log configuration.
type logConfigContextKey struct{}

var logConfigContextKeyValue = logConfigContextKey{} // nolint:gochecknoglobals

// LogConfigToContext creates a new context with per-request logging configuration.
//
// This allows overriding the default log behavior for specific requests.
func LogConfigToContext(
	ctx context.Context,
	logRequestBody bool,
	logResponseBody bool,
	opts ...LogOption,
) context.Context {
	cfg := logConfig{
		LogRequestBody:  logRequestBody,
		LogResponseBody: logResponseBody,
	}

	for _, opt := range opts {
		opt(&cfg)
	}
	return context.WithValue(ctx, logConfigContextKeyValue, cfg)
}

// Log creates a middleware that logs HTTP requests and responses with
// sensible defaults (request body and response body enabled).
//
// All logs are at Debug level unless the response has an error status code.
func Log(logger log.Logger) httpcli.Middleware {
	cfg := &logConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
		CombinedLog:     false,
	}

	return dynamicLogMiddleware(logger, cfg)
}

// LogWithOptions creates a middleware that logs HTTP requests and responses
// with custom configuration.
//
// By default, no body content is logged. Use LogDump, LogHeaders, or other
// options to enable specific logging features.
func LogWithOptions(logger log.Logger, opts ...LogOption) httpcli.Middleware {
	cfg := &logConfig{
		LogRequestBody:  false,
		LogResponseBody: false,
		CombinedLog:     false,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return dynamicLogMiddleware(logger, cfg)
}

// dynamicLogMiddleware creates a middleware that logs HTTP requests and responses
// using the provided logger and configuration.
//
// Per-request configuration from context is merged with the default configuration.
func dynamicLogMiddleware(logger log.Logger, defaultCfg *logConfig) httpcli.Middleware {
	return func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			config := mergeLogConfig(ctx, defaultCfg)

			if config.CombinedLog {
				return combinedLogMiddlewareHandler(ctx, logger, config, next, request)
			}
			return logMiddlewareHandler(ctx, logger, config, next, request)
		})
	}
}

// logMiddlewareHandler handles logging for HTTP requests and responses separately.
//
// Logs the request before execution and the response after execution.
func logMiddlewareHandler(
	ctx context.Context,
	logger log.Logger,
	config logConfig,
	next httpcli.RoundTripper,
	request *httpcli.Request,
) (*httpcli.Response, error) {
	requestFields := []log.Field{
		log.String("method", request.Raw.Method),
		log.String("url", request.Raw.URL.String()),
	}

	if config.LogRequestBody {
		requestFields = append(requestFields, log.ByteString("requestBody", request.Body()))
	}

	logger.Debug(ctx, "http client: request", requestFields...)

	now := time.Now()
	resp, err := next.RoundTrip(ctx, request)

	var responseFields []log.Field

	if config.LogHeadersRequest {
		headers, _ := json.Marshal(request.Raw.Header)
		responseFields = append(responseFields, log.ByteString("requestHeaders", headers))
	}

	if config.LogDumpRequest {
		request.Raw.Body = io.NopCloser(bytes.NewBuffer(request.Body()))
		dumpReq, _ := httputil.DumpRequestOut(request.Raw, true)
		responseFields = append(responseFields, log.ByteString("requestDump", dumpReq))
	}

	if err != nil {
		responseFields = append(responseFields,
			log.Any("error", err),
			log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
		)

		logger.Debug(ctx, "http client: response with error", responseFields...)

		return resp, err
	}

	responseFields = append(responseFields, log.Int("statusCode", resp.StatusCode()))

	if config.LogDumpResponse {
		dumpResp, _ := httputil.DumpResponse(resp.Raw, true)
		responseFields = append(responseFields, log.ByteString("responseDump", dumpResp))
	}

	if config.LogHeadersResponse {
		headers, _ := json.Marshal(resp.Raw.Header)
		responseFields = append(responseFields, log.ByteString("responseHeaders", headers))
	}

	if config.LogResponseBody {
		responseBody, _ := resp.UnsafeBody()
		responseFields = append(responseFields, log.ByteString("responseBody", responseBody))
	}

	responseFields = append(responseFields, log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()))

	logResponseByStatusCode(ctx, logger, responseFields, resp)

	return resp, err
}

// combinedLogMiddlewareHandler handles logging for HTTP requests and responses
// in a single combined log entry.
func combinedLogMiddlewareHandler(
	ctx context.Context,
	logger log.Logger,
	config logConfig,
	next httpcli.RoundTripper,
	request *httpcli.Request,
) (*httpcli.Response, error) {
	var logFields []log.Field
	logFields = append(logFields,
		log.String("method", request.Raw.Method),
		log.String("url", request.Raw.URL.String()),
	)

	if config.LogRequestBody {
		logFields = append(logFields, log.ByteString("requestBody", request.Body()))
	}

	now := time.Now()
	resp, err := next.RoundTrip(ctx, request)

	if config.LogHeadersRequest {
		headers, _ := json.Marshal(request.Raw.Header)
		logFields = append(logFields, log.ByteString("requestHeaders", headers))
	}

	if config.LogDumpRequest {
		request.Raw.Body = io.NopCloser(bytes.NewBuffer(request.Body()))
		dumpReq, _ := httputil.DumpRequestOut(request.Raw, true)
		logFields = append(logFields, log.ByteString("requestDump", dumpReq))
	}

	if err != nil {
		logFields = append(logFields,
			log.Any("error", err),
			log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
		)

		logger.Debug(ctx, "http client: log with error", logFields...)

		return resp, err
	}

	logFields = append(logFields, log.Int("statusCode", resp.StatusCode()))

	if config.LogDumpResponse {
		dumpResp, _ := httputil.DumpResponse(resp.Raw, true)
		logFields = append(logFields, log.ByteString("responseDump", dumpResp))
	}

	if config.LogHeadersResponse {
		headers, _ := json.Marshal(resp.Raw.Header)
		logFields = append(logFields, log.ByteString("responseHeaders", headers))
	}

	if config.LogResponseBody {
		responseBody, _ := resp.UnsafeBody()
		logFields = append(logFields, log.ByteString("responseBody", responseBody))
	}

	logFields = append(logFields, log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()))

	logCombinedByStatusCode(ctx, logger, logFields, resp)

	return resp, err
}

// mergeLogConfig merges per-request log configuration from context with defaults.
//
// Returns the default configuration if no per-request configuration is found.
func mergeLogConfig(ctx context.Context, defaultCfg *logConfig) logConfig {
	configFromContext, ok := ctx.Value(logConfigContextKeyValue).(logConfig)
	if !ok {
		return *defaultCfg
	}

	return configFromContext
}

// logResponseByStatusCode logs response information with appropriate log level
// based on the HTTP status code.
//
// 5xx errors are logged as Error, 4xx as Warn, and others as Debug.
func logResponseByStatusCode(ctx context.Context, logger log.Logger, responseFields []log.Field, resp *httpcli.Response) {
	switch {
	case resp.StatusCode() >= http.StatusInternalServerError:
		logger.Error(ctx, "http client: response", responseFields...)
	case resp.StatusCode() >= http.StatusBadRequest:
		logger.Warn(ctx, "http client: response", responseFields...)
	default:
		logger.Debug(ctx, "http client: response", responseFields...)
	}
}

// logCombinedByStatusCode logs combined request/response information with
// appropriate log level based on the HTTP status code.
//
// 5xx errors are logged as Error, 4xx as Warn, and others as Debug.
func logCombinedByStatusCode(ctx context.Context, logger log.Logger, logFields []log.Field, resp *httpcli.Response) {
	switch {
	case resp.StatusCode() >= http.StatusInternalServerError:
		logger.Error(ctx, "http client: log", logFields...)
	case resp.StatusCode() >= http.StatusBadRequest:
		logger.Warn(ctx, "http client: log", logFields...)
	default:
		logger.Debug(ctx, "http client: log", logFields...)
	}
}

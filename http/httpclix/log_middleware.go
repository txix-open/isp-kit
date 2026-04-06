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

type logConfig struct {
	LogRequestBody  bool
	LogResponseBody bool

	LogDumpRequest  bool
	LogDumpResponse bool

	LogHeadersRequest  bool
	LogHeadersResponse bool

	CombinedLog bool
}

type LogOption func(*logConfig)

func LogDump(dumpRequest bool, dumpResponse bool) LogOption {
	return func(cfg *logConfig) {
		cfg.LogDumpRequest = dumpRequest
		cfg.LogDumpResponse = dumpResponse
	}
}

func LogHeaders(requestHeaders bool, responseHeaders bool) LogOption {
	return func(cfg *logConfig) {
		cfg.LogHeadersRequest = requestHeaders
		cfg.LogHeadersResponse = responseHeaders
	}
}

func LogCombined(combinedLog bool) LogOption {
	return func(cfg *logConfig) {
		cfg.CombinedLog = combinedLog
	}
}

type logConfigContextKey struct{}

var logConfigContextKeyValue = logConfigContextKey{} // nolint:gochecknoglobals

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

func Log(logger log.Logger) httpcli.Middleware {
	cfg := &logConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
		CombinedLog:     false,
	}

	return dynamicLogMiddleware(logger, cfg)
}

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

func mergeLogConfig(ctx context.Context, defaultCfg *logConfig) logConfig {
	configFromContext, ok := ctx.Value(logConfigContextKeyValue).(logConfig)
	if !ok {
		return *defaultCfg
	}

	return configFromContext
}

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

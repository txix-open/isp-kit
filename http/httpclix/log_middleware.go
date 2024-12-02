package httpclix

import (
	"context"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/log"
	"net/http/httputil"
	"time"
)

type logConfig struct {
	LogRequestBody  bool
	LogResponseBody bool

	LogDumpRequest  bool
	LogDumpResponse bool

	LogHeadersRequest  bool
	LogHeadersResponse bool
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

type logConfigContextKey struct{}

var (
	defaultLogConfig = logConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
	}
	logConfigContextKeyValue = logConfigContextKey{}
)

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
	return func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			config := defaultLogConfig
			configFromContext, ok := ctx.Value(logConfigContextKeyValue).(logConfig)
			if ok {
				config = configFromContext
			}

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
				dumpReq, _ := httputil.DumpRequestOut(request.Raw, true)
				responseFields = append(responseFields, log.ByteString("requestDump", dumpReq))
			}

			if err != nil {
				errorFields := append(responseFields,
					log.Any("error", err),
					log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
				)

				logger.Debug(
					ctx,
					"http client: response with error",
					errorFields...,
				)
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
				responseBody, _ := resp.Body()
				responseFields = append(responseFields, log.ByteString("responseBody", responseBody))
			}

			responseFields = append(responseFields, log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()))
			logger.Debug(ctx, "http client: response", responseFields...)

			return resp, err
		})
	}
}

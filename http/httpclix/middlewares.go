package httpclix

import (
	"context"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"github.com/txix-open/isp-kit/requestid"
	"net/http/httputil"
	"time"
)

const (
	RequestIdHeader = "x-request-id"
)

func LogHeadersToContext(ctx context.Context) context.Context {
	cfg, ok := ctx.Value(logConfigContextKeyValue).(logConfig)
	if !ok {
		return context.WithValue(ctx, logConfigContextKeyValue, logConfig{
			LogHeadersRequest:  true,
			LogHeadersResponse: true,
		})
	} else {
		cfg.LogHeadersRequest = true
		cfg.LogHeadersResponse = true
		return context.WithValue(ctx, logConfigContextKeyValue, cfg)
	}
}

func EnableReqRespDump(ctx context.Context) context.Context {
	cfg, ok := ctx.Value(logConfigContextKeyValue).(logConfig)
	if !ok {
		return context.WithValue(ctx, logConfigContextKeyValue, logConfig{
			LogRawRequestBody:  true,
			LogRawResponseBody: true,
		})
	} else {
		cfg.LogRawRequestBody = true
		cfg.LogRawResponseBody = true
		return context.WithValue(ctx, logConfigContextKeyValue, cfg)
	}
}

func RequestId() httpcli.Middleware {
	return func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}

			request.Raw.Header.Set(RequestIdHeader, requestId)
			return next.RoundTrip(ctx, request)
		})
	}
}

type logConfig struct {
	LogRequestBody  bool
	LogResponseBody bool

	LogRawRequestBody  bool
	LogRawResponseBody bool

	LogHeadersRequest  bool
	LogHeadersResponse bool
}

type logConfigContextKey struct{}

var (
	defaultLogConfig = logConfig{
		LogRequestBody:     true,
		LogResponseBody:    true,
		LogRawRequestBody:  false,
		LogRawResponseBody: false,
	}
	logConfigContextKeyValue = logConfigContextKey{}
)

func LogConfigToContext(ctx context.Context, logRequestBody, logResponseBody, logRawRequestBody, logRawResponseBody bool) context.Context {
	return context.WithValue(ctx, logConfigContextKeyValue, logConfig{
		LogRequestBody:     logRequestBody,
		LogResponseBody:    logResponseBody,
		LogRawRequestBody:  logRawRequestBody,
		LogRawResponseBody: logRawResponseBody,
	})
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

			if config.LogHeadersRequest {
				requestFields = append(requestFields, log.Any("header", request.Raw.Header))
			}

			dumpReq, _ := httputil.DumpRequestOut(request.Raw, true)
			requestFields = append(requestFields, log.ByteString("dump request: ", dumpReq))

			if config.LogRequestBody {
				requestFields = append(requestFields, log.ByteString("requestBody", request.Body()))
			}
			logger.Debug(ctx, "http client: request", requestFields...)

			now := time.Now()
			resp, err := next.RoundTrip(ctx, request)
			if err != nil {
				logger.Debug(
					ctx,
					"http client: response with error",
					log.Any("error", err),
					log.Int64("elapsedTimeMs", time.Since(now).Milliseconds()),
				)
				return resp, err
			}

			responseFields := []log.Field{
				log.Int("statusCode", resp.StatusCode()),
			}

			dumpResp, _ := httputil.DumpResponse(resp.Raw, true)
			responseFields = append(responseFields, log.ByteString("dump response: ", dumpResp))

			if config.LogHeadersResponse {
				responseFields = append(responseFields, log.Any("header", resp.Raw.Header))
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

func Metrics(storage *http_metrics.ClientStorage) httpcli.Middleware {
	return func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			endpoint := http_metrics.ClientEndpoint(ctx)
			if endpoint == "" {
				return next.RoundTrip(ctx, request)
			}

			start := time.Now()
			resp, err := next.RoundTrip(ctx, request)
			storage.ObserveDuration(endpoint, time.Since(start))
			return resp, err
		})
	}
}

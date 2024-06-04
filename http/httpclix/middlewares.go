package httpclix

import (
	"context"
	"time"

	"gitlab.txix.ru/isp/isp-kit/http/httpcli"
	"gitlab.txix.ru/isp/isp-kit/log"
	"gitlab.txix.ru/isp/isp-kit/metrics/http_metrics"
	"gitlab.txix.ru/isp/isp-kit/requestid"
)

const (
	RequestIdHeader = "x-request-id"
)

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
}

type logConfigContextKey struct{}

var (
	defaultLogConfig = logConfig{
		LogRequestBody:  true,
		LogResponseBody: true,
	}
	logConfigContextKeyValue = logConfigContextKey{}
)

func LogConfigToContext(ctx context.Context, logRequestBody bool, logResponseBody bool) context.Context {
	return context.WithValue(ctx, logConfigContextKeyValue, logConfig{
		LogRequestBody:  logRequestBody,
		LogResponseBody: logResponseBody,
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

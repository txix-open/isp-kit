package httpclix

import (
	"context"
	"time"

	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/requestid"
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

func Log(logger log.Logger) httpcli.Middleware {
	return func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			logger.Debug(
				ctx,
				"http client: request",
				log.String("method", request.Raw.Method),
				log.String("url", request.Raw.URL.String()),
				log.ByteString("requestBody", request.Body()),
			)

			resp, err := next.RoundTrip(ctx, request)
			if err != nil {
				logger.Debug(ctx, "http client: response", log.Any("error", err))
				return resp, err
			}

			responseBody, _ := resp.Body()
			logger.Debug(
				ctx,
				"http client: response",
				log.Int("statusCode", resp.StatusCode()),
				log.ByteString("responseBody", responseBody),
			)

			return resp, err
		})
	}
}

type MetricStorage interface {
	ObserveDuration(url string, duration time.Duration)
}

func Metrics(storage MetricStorage) httpcli.Middleware {
	return func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			start := time.Now()
			resp, err := next.RoundTrip(ctx, request)
			storage.ObserveDuration(request.Raw.URL.String(), time.Since(start))
			return resp, err
		})
	}
}

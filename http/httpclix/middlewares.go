package httpclix

import (
	"context"
	"net/http/httptrace"
	"time"

	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/http/client_tracing"
	"github.com/txix-open/isp-kit/requestid"
)

func DefaultMiddlewares() []httpcli.Middleware {
	return []httpcli.Middleware{
		RequestId(),
		Metrics(http_metrics.NewClientStorage(metrics.DefaultRegistry)),
		client_tracing.NewConfig().Middleware(),
	}
}

func RequestId() httpcli.Middleware {
	return func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}

			request.Raw.Header.Set(requestid.Header, requestId)
			return next.RoundTrip(ctx, request)
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

			clientTracer := NewClientTracer(storage, endpoint)
			ctx = httptrace.WithClientTrace(ctx, clientTracer.ClientTrace())

			ctx = context.WithValue(ctx, httpcli.ReadingResponseMetricHookKey, clientTracer.ResponseReceived)
			request.Raw = request.Raw.WithContext(ctx)

			start := time.Now()
			resp, err := next.RoundTrip(ctx, request)

			storage.ObserveDuration(endpoint, time.Since(start))
			return resp, err
		})
	}
}

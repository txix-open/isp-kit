package client_tracing

import (
	"context"
	"fmt"
	"net/http/httptrace"

	"github.com/integration-system/isp-kit/http/httpcli"
	"github.com/integration-system/isp-kit/metrics/http_metrics"
	"github.com/integration-system/isp-kit/observability/tracing"
	"github.com/integration-system/isp-kit/observability/tracing/http/semconvutil"
	"github.com/integration-system/isp-kit/requestid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracerName = "isp-kit/observability/httptracing"
)

type Config struct {
	Provider          tracing.TracerProvider
	Propagator        tracing.Propagator
	EnableHttpTracing bool
}

func NewConfig() Config {
	return Config{
		Provider:   tracing.DefaultProvider,
		Propagator: tracing.DefaultPropagator,
	}
}

func (c Config) Middleware() httpcli.Middleware {
	_, isNoop := c.Provider.(tracing.NoopProvider)
	if isNoop {
		return func(next httpcli.RoundTripper) httpcli.RoundTripper {
			return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
				return next.RoundTrip(ctx, request)
			})
		}
	}

	tracer := c.Provider.Tracer(tracerName)
	return func(next httpcli.RoundTripper) httpcli.RoundTripper {
		return httpcli.RoundTripperFunc(func(ctx context.Context, request *httpcli.Request) (*httpcli.Response, error) {
			attributes := semconvutil.HTTPClientRequest(request.Raw)
			attributes = append(attributes, tracing.RequestId.String(requestid.FromContext(ctx)))
			opts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindClient),
			}

			clientEndpoint := http_metrics.ClientEndpoint(ctx)
			if clientEndpoint == "" {
				clientEndpoint = request.Raw.URL.Path
			}
			spanName := fmt.Sprintf("%s %s", request.Raw.Method, clientEndpoint)

			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			if c.EnableHttpTracing {
				otelHttpClientTrace := otelhttptrace.NewClientTrace(ctx, otelhttptrace.WithoutHeaders())
				ctx = httptrace.WithClientTrace(ctx, otelHttpClientTrace)
			}

			c.Propagator.Inject(ctx, propagation.HeaderCarrier(request.Raw.Header))

			resp, err := next.RoundTrip(ctx, request)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
			if resp != nil {
				attributes = append(attributes, semconvutil.HTTPClientResponse(resp.Raw)...)
				span.SetStatus(semconvutil.HTTPClientStatus(resp.StatusCode()))
			}

			span.SetAttributes(attributes...)

			return resp, err
		})
	}
}

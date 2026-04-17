// Package client_tracing provides HTTP client middleware for distributed tracing.
package client_tracing

import (
	"context"
	"fmt"
	"net/http/httptrace"

	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"github.com/txix-open/isp-kit/observability/tracing"
	"github.com/txix-open/isp-kit/observability/tracing/http/semconvutil"
	"github.com/txix-open/isp-kit/requestid"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// tracerName identifies the tracer for HTTP client tracing.
const tracerName = "isp-kit/observability/tracing/http"

// Config holds the configuration for the HTTP client tracing middleware.
type Config struct {
	// Provider is the tracer provider used to create tracers.
	Provider tracing.TracerProvider
	// Propagator is the text map propagator for context propagation.
	Propagator tracing.Propagator
	// EnableHttpTracing enables detailed HTTP client tracing using otelhttptrace.
	EnableHttpTracing bool
}

// NewConfig creates a new Config with default values.
func NewConfig() Config {
	return Config{
		Provider:   tracing.DefaultProvider,
		Propagator: tracing.DefaultPropagator,
	}
}

// Middleware returns an HTTP client middleware that creates spans for outgoing requests.
// It injects trace context into request headers, creates a client span, and records
// response status and errors. If the provider is a no-op, it returns a pass-through middleware.
func (c Config) Middleware() httpcli.Middleware {
	if tracing.IsNoop(c.Provider) {
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
			spanName := fmt.Sprintf("HTTP call %s %s", request.Raw.Method, clientEndpoint)

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
			if resp != nil && resp.Raw != nil {
				attributes = append(attributes, semconvutil.HTTPClientResponse(resp.Raw)...)
				span.SetStatus(semconvutil.HTTPClientStatus(resp.StatusCode()))
			}

			span.SetAttributes(attributes...)

			return resp, err
		})
	}
}

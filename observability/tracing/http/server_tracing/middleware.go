// Package server_tracing provides HTTP server middleware for distributed tracing.
package server_tracing

import (
	"context"
	"fmt"
	"net/http"

	http2 "github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/endpoint/buffer"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/log/logutil"
	"github.com/txix-open/isp-kit/metrics/http_metrics"
	"github.com/txix-open/isp-kit/observability/tracing"
	"github.com/txix-open/isp-kit/observability/tracing/http/semconvutil"
	"github.com/txix-open/isp-kit/requestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

// ReadBytesKey is the attribute key for the total number of bytes read from the request body.
var ReadBytesKey = attribute.Key("http.read_bytes")

// WroteBytesKey is the attribute key for the total number of bytes written to the response.
var WroteBytesKey = attribute.Key("http.wrote_bytes")

// tracerName identifies the tracer for HTTP server tracing.
const tracerName = "isp-kit/observability/tracing/http"

// Config holds the configuration for the HTTP server tracing middleware.
type Config struct {
	// Provider is the tracer provider used to create tracers.
	Provider tracing.TracerProvider
	// Propagator is the text map propagator for context propagation.
	Propagator tracing.Propagator
}

// NewConfig creates a new Config with default values.
func NewConfig() Config {
	return Config{
		Provider:   tracing.DefaultProvider,
		Propagator: tracing.DefaultPropagator,
	}
}

// Middleware returns an HTTP server middleware that creates spans for incoming requests.
// It extracts trace context from request headers, creates a server span, and sets
// appropriate attributes including status code and request details. If the provider
// is a no-op, it returns a pass-through middleware.
func (c Config) Middleware() http2.Middleware {
	if tracing.IsNoop(c.Provider) {
		return func(next http2.HandlerFunc) http2.HandlerFunc {
			return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
				return next(ctx, w, r)
			}
		}
	}

	tracer := c.Provider.Tracer(tracerName)
	return func(next http2.HandlerFunc) http2.HandlerFunc {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			ctx = c.Propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

			attributes := semconvutil.HTTPServerRequest("", r)
			attributes = append(attributes, tracing.RequestId.String(requestid.FromContext(ctx)))
			opts := []trace.SpanStartOption{
				trace.WithSpanKind(trace.SpanKindServer),
			}

			var spanName string
			serverEndpoint := http_metrics.ServerEndpoint(ctx)
			if serverEndpoint != "" {
				attributes = append(attributes, semconv.HTTPRouteKey.String(serverEndpoint))
				spanName = serverEndpoint
			} else {
				spanName = fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			}

			ctx, span := tracer.Start(ctx, spanName, opts...)
			defer span.End()

			var scSource scSource
			buff, isBuffer := w.(*buffer.Buffer)
			if isBuffer {
				scSource = buff
			} else {
				scSource = &writerWrapper{ResponseWriter: w}
			}

			err := next(ctx, scSource, r)

			if isBuffer {
				attributes = append(
					attributes,
					ReadBytesKey.Int(len(buff.RequestBody())),
					WroteBytesKey.Int(len(buff.ResponseBody())),
				)
			}
			attributes = append(
				attributes,
				semconv.HTTPStatusCode(scSource.StatusCode()),
			)
			if err != nil {
				logLevel := logutil.LogLevelForError(err)
				if logLevel == log.ErrorLevel {
					span.RecordError(err)
				}
			}
			span.SetStatus(semconvutil.HTTPServerStatus(scSource.StatusCode()))
			span.SetAttributes(attributes...)

			return err
		}
	}
}

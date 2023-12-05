package server_tracing

import (
	"context"
	"fmt"
	"net/http"

	http2 "github.com/integration-system/isp-kit/http"
	"github.com/integration-system/isp-kit/http/endpoint/buffer"
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/log/logutil"
	"github.com/integration-system/isp-kit/metrics/http_metrics"
	"github.com/integration-system/isp-kit/observability/tracing"
	"github.com/integration-system/isp-kit/observability/tracing/http/semconvutil"
	"github.com/integration-system/isp-kit/requestid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	ReadBytesKey  = attribute.Key("http.read_bytes")  // if anything was read from the request body, the total number of bytes read
	WroteBytesKey = attribute.Key("http.wrote_bytes") // if anything was written to the response writer, the total number of bytes written

	tracerName = "isp-kit/observability/tracing/http"
)

type Config struct {
	Provider   tracing.TracerProvider
	Propagator tracing.Propagator
}

func NewConfig() Config {
	return Config{
		Provider:   tracing.DefaultProvider,
		Propagator: tracing.DefaultPropagator,
	}
}

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

			spanName := ""
			serverEndpoint := http_metrics.ServerEndpoint(ctx)
			if serverEndpoint != "" {
				attributes = append(attributes, semconv.HTTPRouteKey.String(serverEndpoint))
				spanName = serverEndpoint
			} else {
				spanName = fmt.Sprintf("%s %s", r.Method, r.URL.Path)
			}

			//TODO some of endpoint exported
			/*if publicEndpoint {
				opts = append(opts, trace.WithNewRoot())
				// Linking incoming span context if any for public endpoint.
				if s := trace.SpanContextFromContext(ctx); s.IsValid() && s.IsRemote() {
					opts = append(opts, trace.WithLinks(trace.Link{SpanContext: s}))
				}
			}*/

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

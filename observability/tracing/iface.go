package tracing

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	traceapi "go.opentelemetry.io/otel/trace"
)

type TracerProvider = traceapi.TracerProvider

type Propagator = propagation.TextMapPropagator

type Provider interface {
	TracerProvider
	Shutdown(ctx context.Context) error
}

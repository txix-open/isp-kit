package tracing

import (
	"context"

	"go.opentelemetry.io/otel/propagation"
	traceapi "go.opentelemetry.io/otel/trace"
)

// TracerProvider is an alias for OpenTelemetry's TracerProvider interface.
type TracerProvider = traceapi.TracerProvider

// Propagator is an alias for OpenTelemetry's TextMapPropagator interface.
type Propagator = propagation.TextMapPropagator

// Provider extends TracerProvider with a Shutdown method for graceful cleanup.
type Provider interface {
	TracerProvider
	// Shutdown gracefully shuts down the provider, flushing any pending spans.
	Shutdown(ctx context.Context) error
}

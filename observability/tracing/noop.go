package tracing

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

// NoopProvider is a no-op implementation of the Provider interface.
// It provides a tracer that performs no operations, useful when tracing is disabled.
type NoopProvider struct {
	Provider
}

// NewNoopProvider creates a new no-op tracer provider.
func NewNoopProvider() NoopProvider {
	return NoopProvider{}
}

// Tracer returns a no-op tracer that performs no operations.
func (n NoopProvider) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	return noop.Tracer{}
}

// Shutdown returns nil as no resources need to be cleaned up.
func (n NoopProvider) Shutdown(ctx context.Context) error {
	return nil
}

// IsNoop checks if the given TracerProvider is a NoopProvider.
func IsNoop(provider TracerProvider) bool {
	_, isNoop := provider.(NoopProvider)
	return isNoop
}

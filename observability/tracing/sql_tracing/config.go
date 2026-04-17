// Package sql_tracing provides pgx v5 query tracing for distributed tracing.
package sql_tracing

import (
	"github.com/jackc/pgx/v5"
	"github.com/txix-open/isp-kit/observability/tracing"
)

// tracerName identifies the tracer for SQL tracing.
const tracerName = "isp-kit/observability/tracing/sql"

// Config holds the configuration for SQL query tracing.
type Config struct {
	// Provider is the tracer provider used to create tracers.
	Provider tracing.TracerProvider
	// EnableStatement determines whether to include SQL statements in spans.
	EnableStatement bool
	// EnableArgs determines whether to include query arguments in spans.
	EnableArgs bool
}

// NewConfig creates a new Config with default values.
func NewConfig() Config {
	return Config{
		Provider: tracing.DefaultProvider,
	}
}

// QueryTracer returns a pgx QueryTracer for instrumenting database queries.
// It returns a no-op tracer if the provider is a no-op.
func (c Config) QueryTracer() pgx.QueryTracer {
	if tracing.IsNoop(c.Provider) {
		return noop{}
	}

	tracer := c.Provider.Tracer(tracerName)
	return NewTracer(tracer, c)
}

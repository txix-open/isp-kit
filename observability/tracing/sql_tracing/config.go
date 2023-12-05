package sql_tracing

import (
	"github.com/integration-system/isp-kit/observability/tracing"
	"github.com/jackc/pgx/v5"
)

const (
	tracerName = "isp-kit/observability/tracing/sql"
)

type Config struct {
	Provider        tracing.TracerProvider
	EnableStatement bool
	EnableArgs      bool
}

func NewConfig() Config {
	return Config{
		Provider: tracing.DefaultProvider,
	}
}

func (c Config) QueryTracer() pgx.QueryTracer {
	if tracing.IsNoop(c.Provider) {
		return noop{}
	}

	tracer := c.Provider.Tracer(tracerName)
	return NewTracer(tracer, c)
}

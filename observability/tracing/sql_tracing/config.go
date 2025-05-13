package sql_tracing

import (
	"github.com/jackc/pgx/v5"
	"github.com/txix-open/isp-kit/observability/tracing"
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

// nolint:ireturn
func (c Config) QueryTracer() pgx.QueryTracer {
	if tracing.IsNoop(c.Provider) {
		return noop{}
	}

	tracer := c.Provider.Tracer(tracerName)
	return NewTracer(tracer, c)
}

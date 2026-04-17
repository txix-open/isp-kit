package dbx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/txix-open/isp-kit/log"
)

// LogTracer is a pgx QueryTracer that logs queries at debug level.
type LogTracer struct {
	logger log.Logger
}

// NewLogTracer creates a new LogTracer with the provided logger.
func NewLogTracer(logger log.Logger) LogTracer {
	return LogTracer{
		logger: logger,
	}
}

// TraceQueryStart logs the SQL query when a query starts.
func (l LogTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	l.logger.Debug(ctx, "sql: log tracer", log.String("query", data.SQL))
	return ctx
}

// TraceQueryEnd is called when a query completes. Currently a no-op.
func (l LogTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {

}

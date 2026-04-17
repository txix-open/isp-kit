package sql_tracing

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// noop is a no-op implementation of pgx.QueryTracer.
type noop struct{}

// TraceQueryStart returns the context unchanged without creating a span.
func (n noop) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return ctx
}

// TraceQueryEnd performs no operation.
func (n noop) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}

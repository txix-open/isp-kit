package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

// tracers is a slice of pgx QueryTracers that are invoked sequentially.
type tracers []pgx.QueryTracer

// TraceQueryStart invokes all registered tracers when a query starts.
// It passes the context through each tracer and returns the final context.
// nolint:fatcontext
func (t tracers) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	for _, tracer := range t {
		ctx = tracer.TraceQueryStart(ctx, conn, data)
	}
	return ctx
}

// TraceQueryEnd invokes all registered tracers when a query completes.
func (t tracers) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	for _, tracer := range t {
		tracer.TraceQueryEnd(ctx, conn, data)
	}
}

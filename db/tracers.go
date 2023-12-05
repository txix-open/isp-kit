package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type tracers []pgx.QueryTracer

func (t tracers) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	for _, tracer := range t {
		ctx = tracer.TraceQueryStart(ctx, conn, data)
	}
	return ctx
}

func (t tracers) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	for _, tracer := range t {
		tracer.TraceQueryEnd(ctx, conn, data)
	}
}

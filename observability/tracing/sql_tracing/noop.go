package sql_tracing

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type noop struct{}

func (n noop) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	return ctx
}

func (n noop) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {

}

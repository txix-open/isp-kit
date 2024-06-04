package dbx

import (
	"context"

	"github.com/jackc/pgx/v5"
	"gitlab.txix.ru/isp/isp-kit/log"
)

type LogTracer struct {
	logger log.Logger
}

func NewLogTracer(logger log.Logger) LogTracer {
	return LogTracer{
		logger: logger,
	}
}

func (l LogTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	l.logger.Debug(ctx, "sql: log tracer", log.String("query", data.SQL))
	return ctx
}

func (l LogTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {

}

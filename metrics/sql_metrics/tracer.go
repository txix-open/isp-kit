package sql_metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
)

type tracerContextKey int

const (
	startedAtContextKey = tracerContextKey(1)
	labelContextKey     = tracerContextKey(2)
)

type QueryDurationMetrics struct {
	duration *prometheus.SummaryVec
}

func NewTracer(reg *metrics.Registry) QueryDurationMetrics {
	sqlQueryDuration := metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Subsystem:  "sql",
		Name:       "query_duration_ms",
		Help:       "The latencies of sql query",
		Objectives: metrics.DefaultObjectives,
	}, []string{"operation"}))
	return QueryDurationMetrics{
		duration: sqlQueryDuration,
	}
}

func (m QueryDurationMetrics) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	label := ctx.Value(labelContextKey)
	if label == nil {
		return ctx
	}

	return context.WithValue(ctx, startedAtContextKey, time.Now())
}

func (m QueryDurationMetrics) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	startedAt := ctx.Value(startedAtContextKey)
	if startedAt == nil {
		return
	}
	label := fmt.Sprintf("%s", ctx.Value(labelContextKey))

	m.duration.WithLabelValues(label).Observe(float64(time.Since(startedAt.(time.Time)).Milliseconds()))
}

func OperationLabelToContext(ctx context.Context, label string) context.Context {
	return context.WithValue(ctx, labelContextKey, label)
}

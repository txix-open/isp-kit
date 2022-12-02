package dbx

import (
	"context"
	"fmt"
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
)

type PGQContextKey string

const (
	PGQStartedAt = PGQContextKey("pgq_started_at")
	PGQLabel     = PGQContextKey("pgq_label")
)

type QueryDurationMetrics struct {
	registry *prometheus.SummaryVec
}

func NewMetrics() *QueryDurationMetrics {
	sqlQueryDuration := metrics.DefaultRegistry.GetOrRegister(prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Subsystem:  "dbx",
		Name:       "query_duration_ms",
		Help:       "The latencies of sql query",
		Objectives: metrics.DefaultObjectives,
	}, []string{"operation"})).(*prometheus.SummaryVec)

	return &QueryDurationMetrics{
		registry: sqlQueryDuration,
	}
}

func (m *QueryDurationMetrics) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	label := ctx.Value(PGQLabel)
	if label == nil {
		return ctx
	}

	return context.WithValue(ctx, PGQStartedAt, time.Now())
}

func (m *QueryDurationMetrics) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	startedAt := ctx.Value(PGQStartedAt)
	if startedAt == nil {
		return
	}
	label := fmt.Sprintf("%s", ctx.Value(PGQLabel))

	m.registry.WithLabelValues(label).Observe(float64(time.Since(startedAt.(time.Time)).Milliseconds()))
}

func WriteLabelToContext(ctx context.Context, label string) context.Context {
	return context.WithValue(ctx, PGQLabel, label)
}

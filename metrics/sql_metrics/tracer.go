package sql_metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

type tracerContextKey int

const (
	startedAtContextKey = tracerContextKey(1)
	labelContextKey     = tracerContextKey(2)
)

// QueryDurationMetrics collects metrics for SQL query execution, tracking query
// latencies labeled by operation type.
type QueryDurationMetrics struct {
	duration *prometheus.SummaryVec
}

// NewTracer creates a new QueryDurationMetrics instance and registers its metrics
// with the provided registry. The tracer can be assigned to a pgx connection.
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

// TraceQueryStart is called before a query is executed. It extracts the operation
// label from the context and starts timing the query execution. For transaction
// commands (begin, commit, rollback), it appends the command to the operation label.
func (m QueryDurationMetrics) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	label := OperationLabelFromContext(ctx)
	if label == "" {
		return ctx
	}

	if isTransaction(data.SQL) {
		ctx = OperationLabelToContext(ctx, fmt.Sprintf("%s.%s", label, data.SQL))
	}

	return context.WithValue(ctx, startedAtContextKey, time.Now())
}

// TraceQueryEnd is called after a query is executed. It records the query duration
// using the operation label from the context.
func (m QueryDurationMetrics) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	startedAt := ctx.Value(startedAtContextKey)
	if startedAt == nil {
		return
	}
	label := OperationLabelFromContext(ctx)

	duration := time.Since(startedAt.(time.Time)) // nolint:forcetypeassert
	m.duration.WithLabelValues(label).Observe(metrics.Milliseconds(duration))
}

// OperationLabelToContext stores an operation label in the context.
// This label is used to categorize SQL queries in metrics.
func OperationLabelToContext(ctx context.Context, label string) context.Context {
	return context.WithValue(ctx, labelContextKey, label)
}

// OperationLabelFromContext retrieves the operation label from the context.
// Returns an empty string if no label was set.
func OperationLabelFromContext(ctx context.Context) string {
	value, _ := ctx.Value(labelContextKey).(string)
	return value
}

// isTransaction checks if a SQL operation is a transaction command.
func isTransaction(op string) bool {
	return op == "begin" || op == "commit" || op == "rollback"
}

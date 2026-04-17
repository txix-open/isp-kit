// Package sql_metrics provides Prometheus metric collectors for SQL query execution.
// It integrates with the pgx v5 database driver to trace and record query durations.
//
// Example usage:
//
//	tracer := sql_metrics.NewTracer(reg)
//	conn.Config().Tracer = tracer
//
//	// Optionally set operation labels in context:
//	ctx := sql_metrics.OperationLabelToContext(ctx, "users")
//	rows, _ := conn.Query(ctx, "SELECT * FROM users")
package sql_metrics

// Package pgx_metrics provides utilities for registering PostgreSQL connection pool metrics.
// It exposes pgxpool.Stat metrics via Prometheus-compatible collectors.
//
// Example usage:
//
//	pool, _ := pgxpool.New(ctx, dsn)
//	pgx_metrics.Register(reg, pool, "mydb", 10*time.Second)
package pgx_metrics

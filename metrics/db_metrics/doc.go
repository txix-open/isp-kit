// Package db_metrics provides utilities for registering database connection pool metrics.
// It wraps the Prometheus DBStatsCollector to expose PostgreSQL connection pool statistics.
//
// Example usage:
//
//	db, _ := sql.Open("postgres", dsn)
//	db_metrics.Register(reg, db, "mydb")
package db_metrics

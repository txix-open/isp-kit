package db_metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/txix-open/isp-kit/metrics"
)

// Register adds database connection pool statistics to the registry.
// It uses the provided database connection and labels the metrics with the database name.
func Register(reg *metrics.Registry, db *sql.DB, dbName string) {
	metrics.GetOrRegister(reg, collectors.NewDBStatsCollector(db, dbName))
}

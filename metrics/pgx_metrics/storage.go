package pgx_metrics

import (
	"github.com/txix-open/isp-kit/metrics"
)

// Register adds database connection pool statistics to the registry.
// It uses the provided database connection and labels the metrics with the database name.
func Register(reg *metrics.Registry, statsProvider Stater, dbName string) {
	metrics.GetOrRegister(reg, NewCollector(statsProvider, dbName))
}

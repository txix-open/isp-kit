package db_metrics

import (
	"database/sql"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func Register(reg *metrics.Registry, db *sql.DB, dbName string) {
	reg.GetOrRegister(collectors.NewDBStatsCollector(db, dbName))
}

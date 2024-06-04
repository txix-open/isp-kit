package db_metrics

import (
	"database/sql"

	"github.com/prometheus/client_golang/prometheus/collectors"
	"gitlab.txix.ru/isp/isp-kit/metrics"
)

func Register(reg *metrics.Registry, db *sql.DB, dbName string) {
	metrics.GetOrRegister(reg, collectors.NewDBStatsCollector(db, dbName))
}

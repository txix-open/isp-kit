package pgx_metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// statsProvider abstracts pgxpool pool statistics access.
// It exists to allow easier testing and decoupling from pgxpool implementation.
type statsProvider interface {
	Stat() *pgxpool.Stat
}

// poolStatsCollector exports pgxpool connection pool statistics to Prometheus.
//
// It implements prometheus.Collector and exposes both:
//   - instantaneous pool state (gauges)
//   - cumulative pool activity counters (from pgxpool.Stat)
type poolStatsCollector struct {
	statsProvider statsProvider
	dbName        string

	acquiredConns     *prometheus.Desc
	idleConns         *prometheus.Desc
	totalConns        *prometheus.Desc
	maxConns          *prometheus.Desc
	constructingConns *prometheus.Desc
	newConns          *prometheus.Desc

	acquireCount      *prometheus.Desc
	canceledAcquire   *prometheus.Desc
	emptyAcquireCount *prometheus.Desc

	acquireDuration  *prometheus.Desc
	emptyAcquireWait *prometheus.Desc
}

// NewPoolStatsCollector creates a Prometheus collector for pgxpool statistics.
//
// Metrics are labeled with db_name and follow pgxpool naming conventions.
// The collector reads a snapshot of pool state on each scrape.
//
//nolint:funlen
func NewPoolStatsCollector(statsProvider statsProvider, dbName string) prometheus.Collector {
	fq := func(n string) string {
		return "pgxpool_" + n
	}

	return &poolStatsCollector{
		statsProvider: statsProvider,
		dbName:        dbName,

		acquiredConns: prometheus.NewDesc(
			fq("acquired_conns"),
			"Current number of acquired connections",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		idleConns: prometheus.NewDesc(
			fq("idle_conns"),
			"Current number of idle connections",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		totalConns: prometheus.NewDesc(
			fq("total_conns"),
			"Total number of connections in pool",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		maxConns: prometheus.NewDesc(
			fq("max_conns"),
			"Max pool size",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		constructingConns: prometheus.NewDesc(
			fq("constructing_conns"),
			"Connections being created",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		newConns: prometheus.NewDesc(
			fq("new_conns"),
			"New connections being opened",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		acquireCount: prometheus.NewDesc(
			fq("acquire_count"),
			"Total successful acquires",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		canceledAcquire: prometheus.NewDesc(
			fq("canceled_acquire_count"),
			"Canceled acquires",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		emptyAcquireCount: prometheus.NewDesc(
			fq("empty_acquire_count"),
			"Acquires that waited for connection",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		acquireDuration: prometheus.NewDesc(
			fq("acquire_duration_seconds_total"),
			"Total acquire duration",
			nil,
			prometheus.Labels{"db_name": dbName},
		),

		emptyAcquireWait: prometheus.NewDesc(
			fq("empty_acquire_wait_seconds_total"),
			"Total wait time when pool was empty",
			nil,
			prometheus.Labels{"db_name": dbName},
		),
	}
}

// Describe sends all metric descriptors to Prometheus.
// It is required by prometheus.Collector.
func (c *poolStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquiredConns
	ch <- c.idleConns
	ch <- c.totalConns
	ch <- c.maxConns
	ch <- c.constructingConns
	ch <- c.newConns

	ch <- c.acquireCount
	ch <- c.canceledAcquire
	ch <- c.emptyAcquireCount

	ch <- c.acquireDuration
	ch <- c.emptyAcquireWait
}

// Collect gathers a snapshot of pgxpool statistics and exports them as Prometheus metrics.
//
// Gauges represent instantaneous pool state.
// Counters represent cumulative pool activity since pool creation.
//
// Note: duration-related metrics are cumulative values provided by pgxpool.Stat
// and are exported as counters for compatibility with Prometheus conventions.
func (c *poolStatsCollector) Collect(ch chan<- prometheus.Metric) {
	st := c.statsProvider.Stat()
	ch <- prometheus.MustNewConstMetric(c.acquiredConns, prometheus.GaugeValue, float64(st.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.idleConns, prometheus.GaugeValue, float64(st.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.totalConns, prometheus.GaugeValue, float64(st.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.maxConns, prometheus.GaugeValue, float64(st.MaxConns()))
	ch <- prometheus.MustNewConstMetric(c.constructingConns, prometheus.GaugeValue, float64(st.ConstructingConns()))
	ch <- prometheus.MustNewConstMetric(c.newConns, prometheus.GaugeValue, float64(st.NewConnsCount()))

	ch <- prometheus.MustNewConstMetric(c.acquireCount, prometheus.CounterValue, float64(st.AcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.canceledAcquire, prometheus.CounterValue, float64(st.CanceledAcquireCount()))
	ch <- prometheus.MustNewConstMetric(c.emptyAcquireCount, prometheus.CounterValue, float64(st.EmptyAcquireCount()))

	ch <- prometheus.MustNewConstMetric(c.acquireDuration, prometheus.CounterValue, st.AcquireDuration().Seconds())
	ch <- prometheus.MustNewConstMetric(c.emptyAcquireWait, prometheus.CounterValue, st.EmptyAcquireWaitTime().Seconds())
}

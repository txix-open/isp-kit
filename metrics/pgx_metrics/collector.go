package pgx_metrics

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

// Stater abstracts pgxpool pool statistics access.
// It exists to allow easier testing and decoupling from pgxpool implementation.
type Stater interface {
	Stat() *pgxpool.Stat
}

type staterFunc func() *pgxpool.Stat

// Collector exports pgxpool connection pool statistics to Prometheus.
//
// It implements prometheus.Collector and exposes both:
//   - instantaneous pool state (gauges)
//   - cumulative pool activity counters (from pgxpool.Stat)
type Collector struct {
	stater staterFunc

	acquireCount            *prometheus.Desc
	acquireDuration         *prometheus.Desc
	acquiredConns           *prometheus.Desc
	canceledAcquireCount    *prometheus.Desc
	constructingConns       *prometheus.Desc
	emptyAcquireCount       *prometheus.Desc
	emptyAcquireWaitTime    *prometheus.Desc
	idleConns               *prometheus.Desc
	maxConns                *prometheus.Desc
	totalConns              *prometheus.Desc
	newConnsCount           *prometheus.Desc
	maxLifetimeDestroyCount *prometheus.Desc
	maxIdleDestroyCount     *prometheus.Desc
}

// NewCollector creates a Prometheus collector for pgxpool statistics.
//
// Metrics are labeled with db_name and follow pgxpool naming conventions.
// The collector reads a snapshot of pool state on each scrape.
func NewCollector(stater Stater, dbName string) prometheus.Collector {
	fn := func() *pgxpool.Stat { return stater.Stat() }
	return newCollector(fn, dbName)
}

// newCollector is an internal only constructor for a Collecter. It accepts
// a stater which provides a closure for requesting pgxpool.Stat metrics.
// Labels to each metric and may be nil. A label is recommended when an
// application uses more than one pgxpool.Pool to enable differentiation between them.
func newCollector(stater staterFunc, dbName string) *Collector {
	fq := func(n string) string {
		return "pgxpool_" + n
	}
	labels := prometheus.Labels{"db_name": dbName}

	return &Collector{
		stater: stater,

		acquireCount: prometheus.NewDesc(
			fq("acquire_count"),
			"Cumulative count of successful acquires from the pool.",
			nil, labels),
		acquireDuration: prometheus.NewDesc(
			fq("acquire_duration_ns"),
			"Total duration of all successful acquires from the pool in nanoseconds.",
			nil, labels),
		acquiredConns: prometheus.NewDesc(
			fq("acquired_conns"),
			"Number of currently acquired connections in the pool.",
			nil, labels),
		canceledAcquireCount: prometheus.NewDesc(
			fq("canceled_acquire_count"),
			"Cumulative count of acquires from the pool that were canceled by a context.",
			nil, labels),
		constructingConns: prometheus.NewDesc(
			fq("constructing_conns"),
			"Number of conns with construction in progress in the pool.",
			nil, labels),
		emptyAcquireCount: prometheus.NewDesc(
			fq("empty_acquire"),
			"Cumulative count of successful acquires from the pool that waited for a resource to be released or constructed because the pool was empty.",
			nil, labels),
		emptyAcquireWaitTime: prometheus.NewDesc(
			fq("empty_acquire_wait_time_ns"),
			"Cumulative time in nanoseconds waited for successful acquires from the pool for a resource to be released or constructed because the pool was empty.",
			nil, labels),
		idleConns: prometheus.NewDesc(
			fq("idle_conns"),
			"Number of currently idle conns in the pool.",
			nil, labels),
		maxConns: prometheus.NewDesc(
			fq("max_conns"),
			"Maximum size of the pool.",
			nil, labels),
		totalConns: prometheus.NewDesc(
			fq("total_conns"),
			"Total number of resources currently in the pool. The value is the sum of ConstructingConns, AcquiredConns, and IdleConns.",
			nil, labels),
		newConnsCount: prometheus.NewDesc(
			fq("new_conns_count"),
			"Cumulative count of new connections opened.",
			nil, labels),
		maxLifetimeDestroyCount: prometheus.NewDesc(
			fq("max_lifetime_destroy_count"),
			"Cumulative count of connections destroyed because they exceeded MaxConnLifetime.",
			nil, labels),
		maxIdleDestroyCount: prometheus.NewDesc(
			fq("max_idle_destroy_count"),
			"Cumulative count of connections destroyed because they exceeded MaxConnIdleTime.",
			nil, labels),
	}
}

// Describe sends all metric descriptors to Prometheus.
// It is required by prometheus.Collector.
func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect gathers a snapshot of pgxpool statistics and exports them as Prometheus metrics.
//
// Gauges represent instantaneous pool state.
// Counters represent cumulative pool activity since pool creation.
//
// Note: duration-related metrics are cumulative values provided by pgxpool.Stat
// and are exported as counters for compatibility with Prometheus conventions.
func (c *Collector) Collect(metrics chan<- prometheus.Metric) {
	st := c.stater()
	metrics <- prometheus.MustNewConstMetric(
		c.acquireCount,
		prometheus.CounterValue,
		float64(st.AcquireCount()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.acquireDuration,
		prometheus.CounterValue,
		float64(st.AcquireDuration()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.acquiredConns,
		prometheus.GaugeValue,
		float64(st.AcquiredConns()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.canceledAcquireCount,
		prometheus.CounterValue,
		float64(st.CanceledAcquireCount()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.constructingConns,
		prometheus.GaugeValue,
		float64(st.ConstructingConns()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.emptyAcquireCount,
		prometheus.CounterValue,
		float64(st.EmptyAcquireCount()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.emptyAcquireWaitTime,
		prometheus.CounterValue,
		float64(st.EmptyAcquireWaitTime()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.idleConns,
		prometheus.GaugeValue,
		float64(st.IdleConns()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.maxConns,
		prometheus.GaugeValue,
		float64(st.MaxConns()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.totalConns,
		prometheus.GaugeValue,
		float64(st.TotalConns()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.newConnsCount,
		prometheus.CounterValue,
		float64(st.NewConnsCount()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.maxLifetimeDestroyCount,
		prometheus.CounterValue,
		float64(st.MaxLifetimeDestroyCount()),
	)
	metrics <- prometheus.MustNewConstMetric(
		c.maxIdleDestroyCount,
		prometheus.CounterValue,
		float64(st.MaxIdleDestroyCount()),
	)
}

package stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/metrics"
)

type ConsumerStorage struct {
	dialCount         prometheus.Gauge
	fetchCount        prometheus.Gauge
	messageCount      prometheus.Gauge
	messageBytesCount prometheus.Gauge
	rebalanceCount    prometheus.Gauge
	timeoutCount      prometheus.Gauge
	errorCount        prometheus.Gauge

	avgDialTimeDuration prometheus.Observer
	minDialTimeDuration prometheus.Observer
	maxDialTimeDuration prometheus.Observer

	avgReadTimeDuration prometheus.Observer
	minReadTimeDuration prometheus.Observer
	maxReadTimeDuration prometheus.Observer

	avgWaitTimeDuration prometheus.Observer
	minWaitTimeDuration prometheus.Observer
	maxWaitTimeDuration prometheus.Observer

	avgFetchSizeCount prometheus.Gauge
	minFetchSizeCount prometheus.Gauge
	maxFetchSizeCount prometheus.Gauge

	avgFetchBytesCount prometheus.Gauge
	minFetchBytesCount prometheus.Gauge
	maxFetchBytesCount prometheus.Gauge

	offsetCount        prometheus.Gauge
	lagCount           prometheus.Gauge
	queueLengthCount   prometheus.Gauge
	queueCapacityCount prometheus.Gauge
}

// nolint:funlen,promlinter
func NewConsumerStorage(reg *metrics.Registry, consumerId string) *ConsumerStorage {
	s := &ConsumerStorage{
		dialCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_dial_count",
			Help:      "Count of reader dials",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		fetchCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_fetch_count",
			Help:      "Count of reader fetches",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		messageCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_message_count",
			Help:      "Count of reader messages",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		messageBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_message_bytes_count",
			Help:      "Count of reader message bytes",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		rebalanceCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_rebalance_count",
			Help:      "Count of reader rebalances",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		timeoutCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_timeout_count",
			Help:      "Count of reader timeouts",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		errorCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_error_count",
			Help:      "Count of reader errors",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		avgDialTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_dial_time_duration_ms",
			Help:      "The latency of reader average dial time",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		minDialTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "reader_min_dial_time_duration_ms",
			Help:      "The latency of reader minimum dial time",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		maxDialTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "reader_max_dial_time_duration_ms",
			Help:      "The latency of reader maximum dial time",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		avgReadTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_read_time_duration_ms",
			Help:      "The latency of reader average read time",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		minReadTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "reader_min_read_time_duration_ms",
			Help:      "The latency of reader minimum read time",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		maxReadTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "reader_max_read_time_duration_ms",
			Help:      "The latency of reader maximum read time",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		avgWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_wait_time_duration_ms",
			Help:      "The latency of reader average wait time",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		minWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "reader_min_wait_time_duration_ms",
			Help:      "The latency of reader minimum wait time",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		maxWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "reader_max_wait_time_duration_ms",
			Help:      "The latency of reader maximum wait time",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		avgFetchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_fetch_size_count",
			Help:      "Count of reader average fetch size",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		minFetchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_fetch_size_count",
			Help:      "Count of reader minimum fetch size",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		maxFetchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_fetch_size_count",
			Help:      "Count of reader maximum fetch size",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		avgFetchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_fetch_bytes_count",
			Help:      "Count of reader average fetch bytes",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		minFetchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_fetch_bytes_count",
			Help:      "Count of reader minimum fetch bytes",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		maxFetchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_fetch_bytes_count",
			Help:      "Count of reader maximum fetch bytes",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		offsetCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_offset_count",
			Help:      "Count of reader offset",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		lagCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_lag_count",
			Help:      "Count of reader lag",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		queueLengthCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_queue_length_count",
			Help:      "Count of reader queue length",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
		queueCapacityCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_queue_capacity_count",
			Help:      "Count of reader queue capacity",
		}, []string{"consumerId"})).WithLabelValues(consumerId),
	}
	return s
}

func (c *ConsumerStorage) ObserveConsumerDials(dials int64) {
	c.dialCount.Set(float64(dials))
}

func (c *ConsumerStorage) ObserveConsumerFetches(fetches int64) {
	c.fetchCount.Set(float64(fetches))
}

func (c *ConsumerStorage) ObserveConsumerMessages(messages int64) {
	c.messageCount.Set(float64(messages))
}

func (c *ConsumerStorage) ObserveConsumerMessageBytes(messageBytes int64) {
	c.messageBytesCount.Set(float64(messageBytes))
}

func (c *ConsumerStorage) ObserveConsumerRebalances(rebalances int64) {
	c.rebalanceCount.Set(float64(rebalances))
}

func (c *ConsumerStorage) ObserveConsumerTimeouts(timeouts int64) {
	c.timeoutCount.Set(float64(timeouts))
}

func (c *ConsumerStorage) ObserveConsumerError(errors int64) {
	c.errorCount.Set(float64(errors))
}

func (c *ConsumerStorage) ObserveConsumerDialTime(dialTime kafka.DurationStats) {
	c.avgDialTimeDuration.Observe(metrics.Milliseconds(dialTime.Avg))
	c.minDialTimeDuration.Observe(metrics.Milliseconds(dialTime.Min))
	c.maxDialTimeDuration.Observe(metrics.Milliseconds(dialTime.Max))
}

func (c *ConsumerStorage) ObserveConsumerReadTime(readTime kafka.DurationStats) {
	c.avgReadTimeDuration.Observe(metrics.Milliseconds(readTime.Avg))
	c.minReadTimeDuration.Observe(metrics.Milliseconds(readTime.Min))
	c.maxReadTimeDuration.Observe(metrics.Milliseconds(readTime.Max))
}

func (c *ConsumerStorage) ObserveConsumerWaitTime(waitTime kafka.DurationStats) {
	c.avgWaitTimeDuration.Observe(metrics.Milliseconds(waitTime.Avg))
	c.minWaitTimeDuration.Observe(metrics.Milliseconds(waitTime.Min))
	c.maxWaitTimeDuration.Observe(metrics.Milliseconds(waitTime.Max))
}

func (c *ConsumerStorage) ObserveConsumerFetchSize(fetchSize kafka.SummaryStats) {
	c.avgFetchSizeCount.Set(float64(fetchSize.Avg))
	c.minFetchSizeCount.Set(float64(fetchSize.Min))
	c.maxFetchSizeCount.Set(float64(fetchSize.Max))
}

func (c *ConsumerStorage) ObserveConsumerFetchBytes(fetchBytes kafka.SummaryStats) {
	c.avgFetchBytesCount.Set(float64(fetchBytes.Avg))
	c.minFetchBytesCount.Set(float64(fetchBytes.Min))
	c.maxFetchBytesCount.Set(float64(fetchBytes.Max))
}

func (c *ConsumerStorage) ObserveConsumerOffset(offset int64) {
	c.offsetCount.Set(float64(offset))
}

func (c *ConsumerStorage) ObserveConsumerLag(lag int64) {
	c.lagCount.Set(float64(lag))
}

func (c *ConsumerStorage) ObserveConsumerQueueLength(queueLength int64) {
	c.queueLengthCount.Set(float64(queueLength))
}

func (c *ConsumerStorage) ObserveConsumerQueueCapacity(queueCapacity int64) {
	c.queueCapacityCount.Set(float64(queueCapacity))
}

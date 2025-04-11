package stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/metrics"
)

type ConsumerStorage struct {
	dialCount         *prometheus.GaugeVec
	fetchCount        *prometheus.GaugeVec
	messageCount      *prometheus.GaugeVec
	messageBytesCount *prometheus.GaugeVec
	rebalanceCount    *prometheus.GaugeVec
	timeoutCount      *prometheus.GaugeVec
	errorCount        *prometheus.GaugeVec

	avgDialTimeDuration *prometheus.GaugeVec
	minDialTimeDuration *prometheus.GaugeVec
	maxDialTimeDuration *prometheus.GaugeVec

	avgReadTimeDuration *prometheus.GaugeVec
	minReadTimeDuration *prometheus.GaugeVec
	maxReadTimeDuration *prometheus.GaugeVec

	avgWaitTimeDuration *prometheus.GaugeVec
	minWaitTimeDuration *prometheus.GaugeVec
	maxWaitTimeDuration *prometheus.GaugeVec

	avgFetchSizeCount *prometheus.GaugeVec
	minFetchSizeCount *prometheus.GaugeVec
	maxFetchSizeCount *prometheus.GaugeVec

	avgFetchBytesCount *prometheus.GaugeVec
	minFetchBytesCount *prometheus.GaugeVec
	maxFetchBytesCount *prometheus.GaugeVec

	offsetCount        *prometheus.GaugeVec
	lagCount           *prometheus.GaugeVec
	queueLengthCount   *prometheus.GaugeVec
	queueCapacityCount *prometheus.GaugeVec
}

func NewConsumerStorage(reg *metrics.Registry) *ConsumerStorage {
	s := &ConsumerStorage{
		dialCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_dial_count",
			Help:      "Count of reader dials",
		}, []string{"consumerId"})),
		fetchCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_fetch_count",
			Help:      "Count of reader fetches",
		}, []string{"consumerId"})),
		messageCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_message_count",
			Help:      "Count of reader messages",
		}, []string{"consumerId"})),
		messageBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_message_bytes_count",
			Help:      "Count of reader message bytes",
		}, []string{"consumerId"})),
		rebalanceCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_rebalance_count",
			Help:      "Count of reader rebalances",
		}, []string{"consumerId"})),
		timeoutCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_timeout_count",
			Help:      "Count of reader timeouts",
		}, []string{"consumerId"})),
		errorCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_error_count",
			Help:      "Count of reader errors",
		}, []string{"consumerId"})),
		avgDialTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_dial_time_duration_ms",
			Help:      "The latency of reader average dial time",
		}, []string{"consumerId"})),
		minDialTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_dial_time_duration_ms",
			Help:      "The latency of reader minimum dial time",
		}, []string{"consumerId"})),
		maxDialTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_dial_time_duration_ms",
			Help:      "The latency of reader maximum dial time",
		}, []string{"consumerId"})),
		avgReadTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_read_time_duration_ms",
			Help:      "The latency of reader average read time",
		}, []string{"consumerId"})),
		minReadTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_read_time_duration_ms",
			Help:      "The latency of reader minimum read time",
		}, []string{"consumerId"})),
		maxReadTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_read_time_duration_ms",
			Help:      "The latency of reader maximum read time",
		}, []string{"consumerId"})),
		avgWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_wait_time_duration_ms",
			Help:      "The latency of reader average wait time",
		}, []string{"consumerId"})),
		minWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_wait_time_duration_ms",
			Help:      "The latency of reader minimum wait time",
		}, []string{"consumerId"})),
		maxWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_wait_time_duration_ms",
			Help:      "The latency of reader maximum wait time",
		}, []string{"consumerId"})),
		avgFetchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_fetch_size_count",
			Help:      "Count of reader average fetch size",
		}, []string{"consumerId"})),
		minFetchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_fetch_size_count",
			Help:      "Count of reader minimum fetch size",
		}, []string{"consumerId"})),
		maxFetchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_fetch_size_count",
			Help:      "Count of reader maximum fetch size",
		}, []string{"consumerId"})),
		avgFetchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_fetch_bytes_count",
			Help:      "Count of reader average fetch bytes",
		}, []string{"consumerId"})),
		minFetchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_fetch_bytes_count",
			Help:      "Count of reader minimum fetch bytes",
		}, []string{"consumerId"})),
		maxFetchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_fetch_bytes_count",
			Help:      "Count of reader maximum fetch bytes",
		}, []string{"consumerId"})),
		offsetCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_offset_count",
			Help:      "Count of reader offset",
		}, []string{"consumerId"})),
		lagCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_lag_count",
			Help:      "Count of reader lag",
		}, []string{"consumerId"})),
		queueLengthCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_queue_length_count",
			Help:      "Count of reader queue length",
		}, []string{"consumerId"})),
		queueCapacityCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_queue_capacity_count",
			Help:      "Count of reader queue capacity",
		}, []string{"consumerId"})),
	}
	return s
}

func (c *ConsumerStorage) ObserveConsumerDials(consumerId string, dials int64) {
	c.dialCount.WithLabelValues(consumerId).Set(float64(dials))
}

func (c *ConsumerStorage) ObserveConsumerFetches(consumerId string, fetches int64) {
	c.fetchCount.WithLabelValues(consumerId).Set(float64(fetches))
}

func (c *ConsumerStorage) ObserveConsumerMessages(consumerId string, messages int64) {
	c.messageCount.WithLabelValues(consumerId).Set(float64(messages))
}

func (c *ConsumerStorage) ObserveConsumerMessageBytes(consumerId string, messageBytes int64) {
	c.messageBytesCount.WithLabelValues(consumerId).Set(float64(messageBytes))
}

func (c *ConsumerStorage) ObserveConsumerRebalances(consumerId string, rebalances int64) {
	c.rebalanceCount.WithLabelValues(consumerId).Set(float64(rebalances))
}

func (c *ConsumerStorage) ObserveConsumerTimeouts(consumerId string, timeouts int64) {
	c.timeoutCount.WithLabelValues(consumerId).Set(float64(timeouts))
}

func (c *ConsumerStorage) ObserveConsumerError(consumerId string, errors int64) {
	c.errorCount.WithLabelValues(consumerId).Set(float64(errors))
}

func (c *ConsumerStorage) ObserveConsumerDialTime(consumerId string, dialTime kafka.DurationStats) {
	c.avgDialTimeDuration.WithLabelValues(consumerId).Set(metrics.Milliseconds(dialTime.Avg))
	c.minDialTimeDuration.WithLabelValues(consumerId).Set(metrics.Milliseconds(dialTime.Min))
	c.maxDialTimeDuration.WithLabelValues(consumerId).Set(metrics.Milliseconds(dialTime.Max))
}

func (c *ConsumerStorage) ObserveConsumerReadTime(consumerId string, readTime kafka.DurationStats) {
	c.avgReadTimeDuration.WithLabelValues(consumerId).Set(metrics.Milliseconds(readTime.Avg))
	c.minReadTimeDuration.WithLabelValues(consumerId).Set(metrics.Milliseconds(readTime.Min))
	c.maxReadTimeDuration.WithLabelValues(consumerId).Set(metrics.Milliseconds(readTime.Max))
}

func (c *ConsumerStorage) ObserveConsumerWaitTime(consumerId string, waitTime kafka.DurationStats) {
	c.avgWaitTimeDuration.WithLabelValues(consumerId).Set(metrics.Milliseconds(waitTime.Avg))
	c.minWaitTimeDuration.WithLabelValues(consumerId).Set(metrics.Milliseconds(waitTime.Min))
	c.maxWaitTimeDuration.WithLabelValues(consumerId).Set(metrics.Milliseconds(waitTime.Max))
}

func (c *ConsumerStorage) ObserveConsumerFetchSize(consumerId string, fetchSize kafka.SummaryStats) {
	c.avgFetchSizeCount.WithLabelValues(consumerId).Set(float64(fetchSize.Avg))
	c.minFetchSizeCount.WithLabelValues(consumerId).Set(float64(fetchSize.Min))
	c.maxFetchSizeCount.WithLabelValues(consumerId).Set(float64(fetchSize.Max))
}

func (c *ConsumerStorage) ObserveConsumerFetchBytes(consumerId string, fetchBytes kafka.SummaryStats) {
	c.avgFetchBytesCount.WithLabelValues(consumerId).Set(float64(fetchBytes.Avg))
	c.minFetchBytesCount.WithLabelValues(consumerId).Set(float64(fetchBytes.Min))
	c.maxFetchBytesCount.WithLabelValues(consumerId).Set(float64(fetchBytes.Max))
}

func (c *ConsumerStorage) ObserveConsumerOffset(consumerId string, offset int64) {
	c.offsetCount.WithLabelValues(consumerId).Set(float64(offset))
}

func (c *ConsumerStorage) ObserveConsumerLag(consumerId string, lag int64) {
	c.lagCount.WithLabelValues(consumerId).Set(float64(lag))
}

func (c *ConsumerStorage) ObserveConsumerQueueLength(consumerId string, queueLength int64) {
	c.queueLengthCount.WithLabelValues(consumerId).Set(float64(queueLength))
}

func (c *ConsumerStorage) ObserveConsumerQueueCapacity(consumerId string, queueCapacity int64) {
	c.queueCapacityCount.WithLabelValues(consumerId).Set(float64(queueCapacity))
}

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
		}, []string{"consumerId", "topic"})),
		fetchCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_fetch_count",
			Help:      "Count of reader fetches",
		}, []string{"consumerId", "topic"})),
		messageCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_message_count",
			Help:      "Count of reader messages",
		}, []string{"consumerId", "topic"})),
		messageBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_message_bytes_count",
			Help:      "Count of reader message bytes",
		}, []string{"consumerId", "topic"})),
		rebalanceCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_rebalance_count",
			Help:      "Count of reader rebalances",
		}, []string{"consumerId", "topic"})),
		timeoutCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_timeout_count",
			Help:      "Count of reader timeouts",
		}, []string{"consumerId", "topic"})),
		errorCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_error_count",
			Help:      "Count of reader errors",
		}, []string{"consumerId", "topic"})),
		avgDialTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_dial_time_duration_ms",
			Help:      "The latency of reader average dial time",
		}, []string{"consumerId", "topic"})),
		minDialTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_dial_time_duration_ms",
			Help:      "The latency of reader minimum dial time",
		}, []string{"consumerId", "topic"})),
		maxDialTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_dial_time_duration_ms",
			Help:      "The latency of reader maximum dial time",
		}, []string{"consumerId", "topic"})),
		avgReadTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_read_time_duration_ms",
			Help:      "The latency of reader average read time",
		}, []string{"consumerId", "topic"})),
		minReadTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_read_time_duration_ms",
			Help:      "The latency of reader minimum read time",
		}, []string{"consumerId", "topic"})),
		maxReadTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_read_time_duration_ms",
			Help:      "The latency of reader maximum read time",
		}, []string{"consumerId", "topic"})),
		avgWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_wait_time_duration_ms",
			Help:      "The latency of reader average wait time",
		}, []string{"consumerId", "topic"})),
		minWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_wait_time_duration_ms",
			Help:      "The latency of reader minimum wait time",
		}, []string{"consumerId", "topic"})),
		maxWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_wait_time_duration_ms",
			Help:      "The latency of reader maximum wait time",
		}, []string{"consumerId", "topic"})),
		avgFetchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_fetch_size_count",
			Help:      "Count of reader average fetch size",
		}, []string{"consumerId", "topic"})),
		minFetchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_fetch_size_count",
			Help:      "Count of reader minimum fetch size",
		}, []string{"consumerId", "topic"})),
		maxFetchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_fetch_size_count",
			Help:      "Count of reader maximum fetch size",
		}, []string{"consumerId", "topic"})),
		avgFetchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_avg_fetch_bytes_count",
			Help:      "Count of reader average fetch bytes",
		}, []string{"consumerId", "topic"})),
		minFetchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_min_fetch_bytes_count",
			Help:      "Count of reader minimum fetch bytes",
		}, []string{"consumerId", "topic"})),
		maxFetchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_max_fetch_bytes_count",
			Help:      "Count of reader maximum fetch bytes",
		}, []string{"consumerId", "topic"})),
		offsetCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_offset_count",
			Help:      "Count of reader offset",
		}, []string{"consumerId", "topic"})),
		lagCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_lag_count",
			Help:      "Count of reader lag",
		}, []string{"consumerId", "topic"})),
		queueLengthCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_queue_length_count",
			Help:      "Count of reader queue length",
		}, []string{"consumerId", "topic"})),
		queueCapacityCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "reader_queue_capacity_count",
			Help:      "Count of reader queue capacity",
		}, []string{"consumerId", "topic"})),
	}
	return s
}

func (c *ConsumerStorage) ObserveConsumerDials(consumerId, topic string, dials int64) {
	c.dialCount.WithLabelValues(consumerId, topic).Set(float64(dials))
}

func (c *ConsumerStorage) ObserveConsumerFetches(consumerId, topic string, fetches int64) {
	c.fetchCount.WithLabelValues(consumerId, topic).Set(float64(fetches))
}

func (c *ConsumerStorage) ObserveConsumerMessages(consumerId, topic string, messages int64) {
	c.messageCount.WithLabelValues(consumerId, topic).Set(float64(messages))
}

func (c *ConsumerStorage) ObserveConsumerMessageBytes(consumerId, topic string, messageBytes int64) {
	c.messageBytesCount.WithLabelValues(consumerId, topic).Set(float64(messageBytes))
}

func (c *ConsumerStorage) ObserveConsumerRebalances(consumerId, topic string, rebalances int64) {
	c.rebalanceCount.WithLabelValues(consumerId, topic).Set(float64(rebalances))
}

func (c *ConsumerStorage) ObserveConsumerTimeouts(consumerId, topic string, timeouts int64) {
	c.timeoutCount.WithLabelValues(consumerId, topic).Set(float64(timeouts))
}

func (c *ConsumerStorage) ObserveConsumerError(consumerId, topic string, errors int64) {
	c.errorCount.WithLabelValues(consumerId, topic).Set(float64(errors))
}

func (c *ConsumerStorage) ObserveConsumerDialTime(consumerId, topic string, dialTime kafka.DurationStats) {
	c.avgDialTimeDuration.WithLabelValues(consumerId, topic).Set(metrics.Milliseconds(dialTime.Avg))
	c.minDialTimeDuration.WithLabelValues(consumerId, topic).Set(metrics.Milliseconds(dialTime.Min))
	c.maxDialTimeDuration.WithLabelValues(consumerId, topic).Set(metrics.Milliseconds(dialTime.Max))
}

func (c *ConsumerStorage) ObserveConsumerReadTime(consumerId, topic string, readTime kafka.DurationStats) {
	c.avgReadTimeDuration.WithLabelValues(consumerId, topic).Set(metrics.Milliseconds(readTime.Avg))
	c.minReadTimeDuration.WithLabelValues(consumerId, topic).Set(metrics.Milliseconds(readTime.Min))
	c.maxReadTimeDuration.WithLabelValues(consumerId, topic).Set(metrics.Milliseconds(readTime.Max))
}

func (c *ConsumerStorage) ObserveConsumerWaitTime(consumerId, topic string, waitTime kafka.DurationStats) {
	c.avgWaitTimeDuration.WithLabelValues(consumerId, topic).Set(metrics.Milliseconds(waitTime.Avg))
	c.minWaitTimeDuration.WithLabelValues(consumerId, topic).Set(metrics.Milliseconds(waitTime.Min))
	c.maxWaitTimeDuration.WithLabelValues(consumerId, topic).Set(metrics.Milliseconds(waitTime.Max))
}

func (c *ConsumerStorage) ObserveConsumerFetchSize(consumerId, topic string, fetchSize kafka.SummaryStats) {
	c.avgFetchSizeCount.WithLabelValues(consumerId, topic).Set(float64(fetchSize.Avg))
	c.minFetchSizeCount.WithLabelValues(consumerId, topic).Set(float64(fetchSize.Min))
	c.maxFetchSizeCount.WithLabelValues(consumerId, topic).Set(float64(fetchSize.Max))
}

func (c *ConsumerStorage) ObserveConsumerFetchBytes(consumerId, topic string, fetchBytes kafka.SummaryStats) {
	c.avgFetchBytesCount.WithLabelValues(consumerId, topic).Set(float64(fetchBytes.Avg))
	c.minFetchBytesCount.WithLabelValues(consumerId, topic).Set(float64(fetchBytes.Min))
	c.maxFetchBytesCount.WithLabelValues(consumerId, topic).Set(float64(fetchBytes.Max))
}

func (c *ConsumerStorage) ObserveConsumerOffset(consumerId, topic string, offset int64) {
	c.offsetCount.WithLabelValues(consumerId, topic).Set(float64(offset))
}

func (c *ConsumerStorage) ObserveConsumerLag(consumerId, topic string, lag int64) {
	c.lagCount.WithLabelValues(consumerId, topic).Set(float64(lag))
}

func (c *ConsumerStorage) ObserveConsumerQueueLength(consumerId, topic string, queueLength int64) {
	c.queueLengthCount.WithLabelValues(consumerId, topic).Set(float64(queueLength))
}

func (c *ConsumerStorage) ObserveConsumerQueueCapacity(consumerId, topic string, queueCapacity int64) {
	c.queueCapacityCount.WithLabelValues(consumerId, topic).Set(float64(queueCapacity))
}

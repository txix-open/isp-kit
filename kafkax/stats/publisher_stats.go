package stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/metrics"
)

type PublisherStorage struct {
	writeCount        *prometheus.GaugeVec
	messageCount      *prometheus.GaugeVec
	messageBytesCount *prometheus.GaugeVec
	errorCount        *prometheus.GaugeVec
	retriesCount      *prometheus.GaugeVec

	avgBatchTimeDuration *prometheus.GaugeVec
	minBatchTimeDuration *prometheus.GaugeVec
	maxBatchTimeDuration *prometheus.GaugeVec

	avgBatchQueueTimeDuration *prometheus.GaugeVec
	minBatchQueueTimeDuration *prometheus.GaugeVec
	maxBatchQueueTimeDuration *prometheus.GaugeVec

	avgWriteTimeDuration *prometheus.GaugeVec
	minWriteTimeDuration *prometheus.GaugeVec
	maxWriteTimeDuration *prometheus.GaugeVec

	avgWaitTimeDuration *prometheus.GaugeVec
	minWaitTimeDuration *prometheus.GaugeVec
	maxWaitTimeDuration *prometheus.GaugeVec

	avgBatchSizeCount *prometheus.GaugeVec
	minBatchSizeCount *prometheus.GaugeVec
	maxBatchSizeCount *prometheus.GaugeVec

	avgBatchBytesCount *prometheus.GaugeVec
	minBatchBytesCount *prometheus.GaugeVec
	maxBatchBytesCount *prometheus.GaugeVec
}

func NewPublisherStorage(reg *metrics.Registry) *PublisherStorage {
	s := &PublisherStorage{
		writeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_write_count",
			Help:      "Count of writer writes",
		}, []string{"publisherId", "topic"})),
		messageCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_message_count",
			Help:      "Count of writer messages",
		}, []string{"publisherId", "topic"})),
		messageBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_message_bytes_count",
			Help:      "Count of writer messages bytes",
		}, []string{"publisherId", "topic"})),
		errorCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_error_count",
			Help:      "Count of writer errors",
		}, []string{"publisherId", "topic"})),
		retriesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_retries_count",
			Help:      "Count of writer retries",
		}, []string{"publisherId", "topic"})),
		avgBatchTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_time_duration_ms",
			Help:      "The latency of writer average batch time",
		}, []string{"publisherId", "topic"})),
		minBatchTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_time_duration_ms",
			Help:      "The latency of writer minimum batch time",
		}, []string{"publisherId", "topic"})),
		maxBatchTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_time_duration_ms",
			Help:      "The latency of writer maximum batch time",
		}, []string{"publisherId", "topic"})),
		avgBatchQueueTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_queue_time_duration_ms",
			Help:      "The latency of writer average batch queue time",
		}, []string{"publisherId", "topic"})),
		minBatchQueueTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_queue_time_duration_ms",
			Help:      "The latency of writer minimum batch queue time",
		}, []string{"publisherId", "topic"})),
		maxBatchQueueTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_queue_time_duration_ms",
			Help:      "The latency of writer maximum batch queue time",
		}, []string{"publisherId", "topic"})),
		avgWriteTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_write_time_duration_ms",
			Help:      "The latency of writer average write time",
		}, []string{"publisherId", "topic"})),
		minWriteTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_write_time_duration_ms",
			Help:      "The latency of writer minimum write time",
		}, []string{"publisherId", "topic"})),
		maxWriteTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_write_time_duration_ms",
			Help:      "The latency of writer maximum write time",
		}, []string{"publisherId", "topic"})),
		avgWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_wait_time_duration_ms",
			Help:      "The latency of writer average wait time",
		}, []string{"publisherId", "topic"})),
		minWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_wait_time_duration_ms",
			Help:      "The latency of writer minimum wait time",
		}, []string{"publisherId", "topic"})),
		maxWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_wait_time_duration_ms",
			Help:      "The latency of writer maximum wait time",
		}, []string{"publisherId", "topic"})),
		avgBatchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_size_count",
			Help:      "Count of writer average batch size",
		}, []string{"publisherId", "topic"})),
		minBatchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_size_count",
			Help:      "Count of writer minimum batch size",
		}, []string{"publisherId", "topic"})),
		maxBatchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_size_count",
			Help:      "Count of writer maximum batch size",
		}, []string{"publisherId", "topic"})),
		avgBatchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_bytes_count",
			Help:      "Count of writer average batch bytes",
		}, []string{"publisherId", "topic"})),
		minBatchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_bytes_count",
			Help:      "Count of writer minimum batch bytes",
		}, []string{"publisherId", "topic"})),
		maxBatchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_bytes_count",
			Help:      "Count of writer maximum batch bytes",
		}, []string{"publisherId", "topic"})),
	}
	return s
}

func (p *PublisherStorage) ObservePublisherWrites(publisherId, topic string, writes int64) {
	p.writeCount.WithLabelValues(publisherId, topic).Set(float64(writes))
}

func (p *PublisherStorage) ObservePublisherMessages(publisherId, topic string, messages int64) {
	p.messageCount.WithLabelValues(publisherId, topic).Set(float64(messages))
}

func (p *PublisherStorage) ObservePublisherMessageBytes(publisherId, topic string, messageBytes int64) {
	p.messageBytesCount.WithLabelValues(publisherId, topic).Set(float64(messageBytes))
}

func (p *PublisherStorage) ObservePublisherErrors(publisherId, topic string, errors int64) {
	p.errorCount.WithLabelValues(publisherId, topic).Set(float64(errors))
}

func (p *PublisherStorage) ObservePublisherRetries(publisherId, topic string, retries int64) {
	p.retriesCount.WithLabelValues(publisherId, topic).Set(float64(retries))
}

func (p *PublisherStorage) ObserveConsumerBatchTime(publisherId, topic string, batchTime kafka.DurationStats) {
	p.avgBatchTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(batchTime.Avg))
	p.minBatchTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(batchTime.Min))
	p.maxBatchTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(batchTime.Max))
}

func (p *PublisherStorage) ObserveConsumerBatchQueueTime(publisherId, topic string, batchQueueTime kafka.DurationStats) {
	p.avgBatchQueueTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(batchQueueTime.Avg))
	p.minBatchQueueTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(batchQueueTime.Min))
	p.maxBatchQueueTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(batchQueueTime.Max))
}

func (p *PublisherStorage) ObserveConsumerWriteTime(publisherId, topic string, writeTime kafka.DurationStats) {
	p.avgWriteTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(writeTime.Avg))
	p.minWriteTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(writeTime.Min))
	p.maxWriteTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(writeTime.Max))
}

func (p *PublisherStorage) ObserveConsumerWaitTime(publisherId, topic string, waitTime kafka.DurationStats) {
	p.avgWaitTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(waitTime.Avg))
	p.minWaitTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(waitTime.Min))
	p.maxWaitTimeDuration.WithLabelValues(publisherId, topic).Set(metrics.Milliseconds(waitTime.Max))
}

func (p *PublisherStorage) ObserveConsumerBatchSize(publisherId, topic string, batchSize kafka.SummaryStats) {
	p.avgBatchSizeCount.WithLabelValues(publisherId, topic).Set(float64(batchSize.Avg))
	p.minBatchSizeCount.WithLabelValues(publisherId, topic).Set(float64(batchSize.Min))
	p.maxBatchSizeCount.WithLabelValues(publisherId, topic).Set(float64(batchSize.Max))
}

func (p *PublisherStorage) ObserveConsumerBatchBytes(publisherId, topic string, batchBytes kafka.SummaryStats) {
	p.avgBatchBytesCount.WithLabelValues(publisherId, topic).Set(float64(batchBytes.Avg))
	p.minBatchBytesCount.WithLabelValues(publisherId, topic).Set(float64(batchBytes.Min))
	p.maxBatchBytesCount.WithLabelValues(publisherId, topic).Set(float64(batchBytes.Max))
}

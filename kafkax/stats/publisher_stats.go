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
		}, []string{"publisherId"})),
		messageCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_message_count",
			Help:      "Count of writer messages",
		}, []string{"publisherId"})),
		messageBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_message_bytes_count",
			Help:      "Count of writer messages bytes",
		}, []string{"publisherId"})),
		errorCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_error_count",
			Help:      "Count of writer errors",
		}, []string{"publisherId"})),
		retriesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_retries_count",
			Help:      "Count of writer retries",
		}, []string{"publisherId"})),
		avgBatchTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_time_duration_ms",
			Help:      "The latency of writer average batch time",
		}, []string{"publisherId"})),
		minBatchTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_time_duration_ms",
			Help:      "The latency of writer minimum batch time",
		}, []string{"publisherId"})),
		maxBatchTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_time_duration_ms",
			Help:      "The latency of writer maximum batch time",
		}, []string{"publisherId"})),
		avgBatchQueueTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_queue_time_duration_ms",
			Help:      "The latency of writer average batch queue time",
		}, []string{"publisherId"})),
		minBatchQueueTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_queue_time_duration_ms",
			Help:      "The latency of writer minimum batch queue time",
		}, []string{"publisherId"})),
		maxBatchQueueTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_queue_time_duration_ms",
			Help:      "The latency of writer maximum batch queue time",
		}, []string{"publisherId"})),
		avgWriteTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_write_time_duration_ms",
			Help:      "The latency of writer average write time",
		}, []string{"publisherId"})),
		minWriteTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_write_time_duration_ms",
			Help:      "The latency of writer minimum write time",
		}, []string{"publisherId"})),
		maxWriteTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_write_time_duration_ms",
			Help:      "The latency of writer maximum write time",
		}, []string{"publisherId"})),
		avgWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_wait_time_duration_ms",
			Help:      "The latency of writer average wait time",
		}, []string{"publisherId"})),
		minWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_wait_time_duration_ms",
			Help:      "The latency of writer minimum wait time",
		}, []string{"publisherId"})),
		maxWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_wait_time_duration_ms",
			Help:      "The latency of writer maximum wait time",
		}, []string{"publisherId"})),
		avgBatchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_size_count",
			Help:      "Count of writer average batch size",
		}, []string{"publisherId"})),
		minBatchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_size_count",
			Help:      "Count of writer minimum batch size",
		}, []string{"publisherId"})),
		maxBatchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_size_count",
			Help:      "Count of writer maximum batch size",
		}, []string{"publisherId"})),
		avgBatchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_bytes_count",
			Help:      "Count of writer average batch bytes",
		}, []string{"publisherId"})),
		minBatchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_bytes_count",
			Help:      "Count of writer minimum batch bytes",
		}, []string{"publisherId"})),
		maxBatchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_bytes_count",
			Help:      "Count of writer maximum batch bytes",
		}, []string{"publisherId"})),
	}
	return s
}

func (p *PublisherStorage) ObservePublisherWrites(publisherId string, writes int64) {
	p.writeCount.WithLabelValues(publisherId).Set(float64(writes))
}

func (p *PublisherStorage) ObservePublisherMessages(publisherId string, messages int64) {
	p.messageCount.WithLabelValues(publisherId).Set(float64(messages))
}

func (p *PublisherStorage) ObservePublisherMessageBytes(publisherId string, messageBytes int64) {
	p.messageBytesCount.WithLabelValues(publisherId).Set(float64(messageBytes))
}

func (p *PublisherStorage) ObservePublisherErrors(publisherId string, errors int64) {
	p.errorCount.WithLabelValues(publisherId).Set(float64(errors))
}

func (p *PublisherStorage) ObservePublisherRetries(publisherId string, retries int64) {
	p.retriesCount.WithLabelValues(publisherId).Set(float64(retries))
}

func (p *PublisherStorage) ObserveConsumerBatchTime(publisherId string, batchTime kafka.DurationStats) {
	p.avgBatchTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(batchTime.Avg))
	p.minBatchTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(batchTime.Min))
	p.maxBatchTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(batchTime.Max))
}

func (p *PublisherStorage) ObserveConsumerBatchQueueTime(publisherId string, batchQueueTime kafka.DurationStats) {
	p.avgBatchQueueTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(batchQueueTime.Avg))
	p.minBatchQueueTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(batchQueueTime.Min))
	p.maxBatchQueueTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(batchQueueTime.Max))
}

func (p *PublisherStorage) ObserveConsumerWriteTime(publisherId string, writeTime kafka.DurationStats) {
	p.avgWriteTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(writeTime.Avg))
	p.minWriteTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(writeTime.Min))
	p.maxWriteTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(writeTime.Max))
}

func (p *PublisherStorage) ObserveConsumerWaitTime(publisherId string, waitTime kafka.DurationStats) {
	p.avgWaitTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(waitTime.Avg))
	p.minWaitTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(waitTime.Min))
	p.maxWaitTimeDuration.WithLabelValues(publisherId).Set(metrics.Milliseconds(waitTime.Max))
}

func (p *PublisherStorage) ObserveConsumerBatchSize(publisherId string, batchSize kafka.SummaryStats) {
	p.avgBatchSizeCount.WithLabelValues(publisherId).Set(float64(batchSize.Avg))
	p.minBatchSizeCount.WithLabelValues(publisherId).Set(float64(batchSize.Min))
	p.maxBatchSizeCount.WithLabelValues(publisherId).Set(float64(batchSize.Max))
}

func (p *PublisherStorage) ObserveConsumerBatchBytes(publisherId string, batchBytes kafka.SummaryStats) {
	p.avgBatchBytesCount.WithLabelValues(publisherId).Set(float64(batchBytes.Avg))
	p.minBatchBytesCount.WithLabelValues(publisherId).Set(float64(batchBytes.Min))
	p.maxBatchBytesCount.WithLabelValues(publisherId).Set(float64(batchBytes.Max))
}

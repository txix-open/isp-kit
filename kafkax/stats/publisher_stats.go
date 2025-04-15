package stats

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/metrics"
)

type PublisherStorage struct {
	writeCount        prometheus.Gauge
	messageCount      prometheus.Gauge
	messageBytesCount prometheus.Gauge
	errorCount        prometheus.Gauge
	retriesCount      prometheus.Gauge

	avgBatchTimeDuration prometheus.Observer
	minBatchTimeDuration prometheus.Observer
	maxBatchTimeDuration prometheus.Observer

	avgBatchQueueTimeDuration prometheus.Observer
	minBatchQueueTimeDuration prometheus.Observer
	maxBatchQueueTimeDuration prometheus.Observer

	avgWriteTimeDuration prometheus.Observer
	minWriteTimeDuration prometheus.Observer
	maxWriteTimeDuration prometheus.Observer

	avgWaitTimeDuration prometheus.Observer
	minWaitTimeDuration prometheus.Observer
	maxWaitTimeDuration prometheus.Observer

	avgBatchSizeCount prometheus.Gauge
	minBatchSizeCount prometheus.Gauge
	maxBatchSizeCount prometheus.Gauge

	avgBatchBytesCount prometheus.Gauge
	minBatchBytesCount prometheus.Gauge
	maxBatchBytesCount prometheus.Gauge
}

func NewPublisherStorage(reg *metrics.Registry, publisherId string) *PublisherStorage {
	s := &PublisherStorage{
		writeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_write_count",
			Help:      "Count of writer writes",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		messageCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_message_count",
			Help:      "Count of writer messages",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		messageBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_message_bytes_count",
			Help:      "Count of writer messages bytes",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		errorCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_error_count",
			Help:      "Count of writer errors",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		retriesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_retries_count",
			Help:      "Count of writer retries",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		avgBatchTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_time_duration_ms",
			Help:      "The latency of writer average batch time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		minBatchTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_time_duration_ms",
			Help:      "The latency of writer minimum batch time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		maxBatchTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_time_duration_ms",
			Help:      "The latency of writer maximum batch time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		avgBatchQueueTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_queue_time_duration_ms",
			Help:      "The latency of writer average batch queue time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		minBatchQueueTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_queue_time_duration_ms",
			Help:      "The latency of writer minimum batch queue time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		maxBatchQueueTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_queue_time_duration_ms",
			Help:      "The latency of writer maximum batch queue time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		avgWriteTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_write_time_duration_ms",
			Help:      "The latency of writer average write time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		minWriteTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_min_write_time_duration_ms",
			Help:      "The latency of writer minimum write time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		maxWriteTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_max_write_time_duration_ms",
			Help:      "The latency of writer maximum write time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		avgWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_wait_time_duration_ms",
			Help:      "The latency of writer average wait time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		minWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_min_wait_time_duration_ms",
			Help:      "The latency of writer minimum wait time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		maxWaitTimeDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem: "kafka",
			Name:      "writer_max_wait_time_duration_ms",
			Help:      "The latency of writer maximum wait time",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		avgBatchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_size_count",
			Help:      "Count of writer average batch size",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		minBatchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_size_count",
			Help:      "Count of writer minimum batch size",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		maxBatchSizeCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_size_count",
			Help:      "Count of writer maximum batch size",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		avgBatchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_avg_batch_bytes_count",
			Help:      "Count of writer average batch bytes",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		minBatchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_min_batch_bytes_count",
			Help:      "Count of writer minimum batch bytes",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
		maxBatchBytesCount: metrics.GetOrRegister(reg, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Subsystem: "kafka",
			Name:      "writer_max_batch_bytes_count",
			Help:      "Count of writer maximum batch bytes",
		}, []string{"publisherId"})).WithLabelValues(publisherId),
	}
	return s
}

func (p *PublisherStorage) ObservePublisherWrites(writes int64) {
	p.writeCount.Set(float64(writes))
}

func (p *PublisherStorage) ObservePublisherMessages(messages int64) {
	p.messageCount.Set(float64(messages))
}

func (p *PublisherStorage) ObservePublisherMessageBytes(messageBytes int64) {
	p.messageBytesCount.Set(float64(messageBytes))
}

func (p *PublisherStorage) ObservePublisherErrors(errors int64) {
	p.errorCount.Set(float64(errors))
}

func (p *PublisherStorage) ObservePublisherRetries(retries int64) {
	p.retriesCount.Set(float64(retries))
}

func (p *PublisherStorage) ObserveConsumerBatchTime(batchTime kafka.DurationStats) {
	p.avgBatchTimeDuration.Observe(metrics.Milliseconds(batchTime.Avg))
	p.minBatchTimeDuration.Observe(metrics.Milliseconds(batchTime.Min))
	p.maxBatchTimeDuration.Observe(metrics.Milliseconds(batchTime.Max))
}

func (p *PublisherStorage) ObserveConsumerBatchQueueTime(batchQueueTime kafka.DurationStats) {
	p.avgBatchQueueTimeDuration.Observe(metrics.Milliseconds(batchQueueTime.Avg))
	p.minBatchQueueTimeDuration.Observe(metrics.Milliseconds(batchQueueTime.Min))
	p.maxBatchQueueTimeDuration.Observe(metrics.Milliseconds(batchQueueTime.Max))
}

func (p *PublisherStorage) ObserveConsumerWriteTime(writeTime kafka.DurationStats) {
	p.avgWriteTimeDuration.Observe(metrics.Milliseconds(writeTime.Avg))
	p.minWriteTimeDuration.Observe(metrics.Milliseconds(writeTime.Min))
	p.maxWriteTimeDuration.Observe(metrics.Milliseconds(writeTime.Max))
}

func (p *PublisherStorage) ObserveConsumerWaitTime(waitTime kafka.DurationStats) {
	p.avgWaitTimeDuration.Observe(metrics.Milliseconds(waitTime.Avg))
	p.minWaitTimeDuration.Observe(metrics.Milliseconds(waitTime.Min))
	p.maxWaitTimeDuration.Observe(metrics.Milliseconds(waitTime.Max))
}

func (p *PublisherStorage) ObserveConsumerBatchSize(batchSize kafka.SummaryStats) {
	p.avgBatchSizeCount.Set(float64(batchSize.Avg))
	p.minBatchSizeCount.Set(float64(batchSize.Min))
	p.maxBatchSizeCount.Set(float64(batchSize.Max))
}

func (p *PublisherStorage) ObserveConsumerBatchBytes(batchBytes kafka.SummaryStats) {
	p.avgBatchBytesCount.Set(float64(batchBytes.Avg))
	p.minBatchBytesCount.Set(float64(batchBytes.Min))
	p.maxBatchBytesCount.Set(float64(batchBytes.Max))
}

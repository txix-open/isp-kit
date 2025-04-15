package publisher

import (
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/stats"
	"github.com/txix-open/isp-kit/metrics"
)

type MetricStorage interface {
	ObservePublisherWrites(writes int64)
	ObservePublisherMessages(messages int64)
	ObservePublisherMessageBytes(messageBytes int64)
	ObservePublisherErrors(errors int64)
	ObservePublisherRetries(retries int64)

	ObserveConsumerBatchTime(batchTime kafka.DurationStats)
	ObserveConsumerBatchQueueTime(batchQueueTime kafka.DurationStats)
	ObserveConsumerWriteTime(writeTime kafka.DurationStats)
	ObserveConsumerWaitTime(waitTime kafka.DurationStats)
	ObserveConsumerBatchSize(batchSize kafka.SummaryStats)
	ObserveConsumerBatchBytes(batchBytes kafka.SummaryStats)
}

type Metrics struct {
	timer     *time.Ticker
	writer    *kafka.Writer
	storage   MetricStorage
	closeChan chan struct{}
}

func NewMetrics(sendMetricPeriod time.Duration, writer *kafka.Writer, publisherId string) *Metrics {
	return &Metrics{
		closeChan: make(chan struct{}),
		writer:    writer,
		timer:     time.NewTicker(sendMetricPeriod),
		storage:   stats.NewPublisherStorage(metrics.DefaultRegistry, publisherId),
	}
}

func (m *Metrics) Send(stats kafka.WriterStats) {
	m.storage.ObservePublisherWrites(stats.Writes)
	m.storage.ObservePublisherMessages(stats.Messages)
	m.storage.ObservePublisherMessageBytes(stats.Bytes)
	m.storage.ObservePublisherErrors(stats.Errors)
	m.storage.ObservePublisherRetries(stats.Retries)

	m.storage.ObserveConsumerBatchTime(stats.BatchTime)
	m.storage.ObserveConsumerBatchQueueTime(stats.BatchQueueTime)
	m.storage.ObserveConsumerWriteTime(stats.WriteTime)
	m.storage.ObserveConsumerWaitTime(stats.WaitTime)
	m.storage.ObserveConsumerBatchSize(stats.BatchSize)
	m.storage.ObserveConsumerBatchBytes(stats.BatchBytes)
}

func (m *Metrics) Run() {
	defer m.Send(m.writer.Stats())
	for {
		select {
		case <-m.closeChan:
			return
		case <-m.timer.C:
			m.Send(m.writer.Stats())
		}
	}
}

func (m *Metrics) Close() {
	m.timer.Stop()
	m.closeChan <- struct{}{}
	close(m.closeChan)
}

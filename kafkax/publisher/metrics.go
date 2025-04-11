package publisher

import (
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/stats"
	"github.com/txix-open/isp-kit/metrics"
)

type MetricStorage interface {
	ObservePublisherWrites(publisherId string, writes int64)
	ObservePublisherMessages(publisherId string, messages int64)
	ObservePublisherMessageBytes(publisherId string, messageBytes int64)
	ObservePublisherErrors(publisherId string, errors int64)
	ObservePublisherRetries(publisherId string, retries int64)

	ObserveConsumerBatchTime(publisherId string, batchTime kafka.DurationStats)
	ObserveConsumerBatchQueueTime(publisherId string, batchQueueTime kafka.DurationStats)
	ObserveConsumerWriteTime(publisherId string, writeTime kafka.DurationStats)
	ObserveConsumerWaitTime(publisherId string, waitTime kafka.DurationStats)
	ObserveConsumerBatchSize(publisherId string, batchSize kafka.SummaryStats)
	ObserveConsumerBatchBytes(publisherId string, batchBytes kafka.SummaryStats)
}

type Metrics struct {
	isSend    bool
	timer     *time.Ticker
	storage   MetricStorage
	closeChan chan struct{}
}

func NewMetrics(isSend bool, sendMetricPeriod time.Duration) Metrics {
	m := Metrics{
		isSend:    isSend,
		closeChan: make(chan struct{}),
	}

	if isSend {
		m.timer = time.NewTicker(sendMetricPeriod)
		m.storage = stats.NewPublisherStorage(metrics.DefaultRegistry)
	}

	return m
}

func (m *Metrics) IsSend() bool {
	return m.isSend
}

func (m *Metrics) Send(stats kafka.WriterStats) {
	m.storage.ObservePublisherWrites(stats.ClientID, stats.Writes)
	m.storage.ObservePublisherMessages(stats.ClientID, stats.Messages)
	m.storage.ObservePublisherMessageBytes(stats.ClientID, stats.Bytes)
	m.storage.ObservePublisherErrors(stats.ClientID, stats.Errors)
	m.storage.ObservePublisherRetries(stats.ClientID, stats.Retries)

	m.storage.ObserveConsumerBatchTime(stats.ClientID, stats.BatchTime)
	m.storage.ObserveConsumerBatchQueueTime(stats.ClientID, stats.BatchQueueTime)
	m.storage.ObserveConsumerWriteTime(stats.ClientID, stats.WriteTime)
	m.storage.ObserveConsumerWaitTime(stats.ClientID, stats.WaitTime)
	m.storage.ObserveConsumerBatchSize(stats.ClientID, stats.BatchSize)
	m.storage.ObserveConsumerBatchBytes(stats.ClientID, stats.BatchBytes)
}

func (m *Metrics) Stop() {
	if m.isSend {
		m.timer.Stop()
	}
}

func (m *Metrics) Close() {
	if m.isSend {
		m.closeChan <- struct{}{}
	}

	close(m.closeChan)
}

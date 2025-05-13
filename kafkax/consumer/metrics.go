package consumer

import (
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/stats"
	"github.com/txix-open/isp-kit/metrics"
)

// nolint:interfacebloat
type MetricStorage interface {
	ObserveConsumerDials(dials int64)
	ObserveConsumerFetches(fetches int64)
	ObserveConsumerMessages(messages int64)
	ObserveConsumerMessageBytes(messageBytes int64)
	ObserveConsumerRebalances(rebalances int64)
	ObserveConsumerTimeouts(timeouts int64)
	ObserveConsumerError(errors int64)

	ObserveConsumerDialTime(dialTime kafka.DurationStats)
	ObserveConsumerReadTime(readTime kafka.DurationStats)
	ObserveConsumerWaitTime(waitTime kafka.DurationStats)
	ObserveConsumerFetchSize(fetchSize kafka.SummaryStats)
	ObserveConsumerFetchBytes(fetchBytes kafka.SummaryStats)

	ObserveConsumerOffset(offset int64)
	ObserveConsumerLag(lag int64)
	ObserveConsumerQueueLength(queueLength int64)
	ObserveConsumerQueueCapacity(queueCapacity int64)
}

type Metrics struct {
	timer     *time.Ticker
	reader    *kafka.Reader
	storage   MetricStorage
	closeChan chan struct{}
}

func NewMetrics(sendMetricPeriod time.Duration, reader *kafka.Reader, consumerId string) *Metrics {
	return &Metrics{
		closeChan: make(chan struct{}),
		reader:    reader,
		timer:     time.NewTicker(sendMetricPeriod),
		storage:   stats.NewConsumerStorage(metrics.DefaultRegistry, consumerId),
	}
}

func (m *Metrics) Send(stats kafka.ReaderStats) {
	m.storage.ObserveConsumerDials(stats.Dials)
	m.storage.ObserveConsumerFetches(stats.Fetches)
	m.storage.ObserveConsumerMessages(stats.Messages)
	m.storage.ObserveConsumerMessageBytes(stats.Bytes)
	m.storage.ObserveConsumerRebalances(stats.Rebalances)
	m.storage.ObserveConsumerTimeouts(stats.Timeouts)
	m.storage.ObserveConsumerError(stats.Errors)

	m.storage.ObserveConsumerDialTime(stats.DialTime)
	m.storage.ObserveConsumerReadTime(stats.ReadTime)
	m.storage.ObserveConsumerWaitTime(stats.WaitTime)
	m.storage.ObserveConsumerFetchSize(stats.FetchSize)
	m.storage.ObserveConsumerFetchBytes(stats.FetchBytes)

	m.storage.ObserveConsumerOffset(stats.Offset)
	m.storage.ObserveConsumerLag(stats.Lag)
	m.storage.ObserveConsumerQueueLength(stats.QueueLength)
	m.storage.ObserveConsumerQueueCapacity(stats.QueueCapacity)
}

func (m *Metrics) Run() {
	defer m.Send(m.reader.Stats())
	for {
		select {
		case <-m.closeChan:
			return
		case <-m.timer.C:
			m.Send(m.reader.Stats())
		}
	}
}

func (m *Metrics) Close() {
	m.timer.Stop()
	m.closeChan <- struct{}{}
	close(m.closeChan)
}

package consumer

import (
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/stats"
	"github.com/txix-open/isp-kit/metrics"
)

type MetricStorage interface {
	ObserveConsumerDials(consumerId string, dials int64)
	ObserveConsumerFetches(consumerId string, fetches int64)
	ObserveConsumerMessages(consumerId string, messages int64)
	ObserveConsumerMessageBytes(consumerId string, messageBytes int64)
	ObserveConsumerRebalances(consumerId string, rebalances int64)
	ObserveConsumerTimeouts(consumerId string, timeouts int64)
	ObserveConsumerError(consumerId string, errors int64)

	ObserveConsumerDialTime(consumerId string, dialTime kafka.DurationStats)
	ObserveConsumerReadTime(consumerId string, readTime kafka.DurationStats)
	ObserveConsumerWaitTime(consumerId string, waitTime kafka.DurationStats)
	ObserveConsumerFetchSize(consumerId string, fetchSize kafka.SummaryStats)
	ObserveConsumerFetchBytes(consumerId string, fetchBytes kafka.SummaryStats)

	ObserveConsumerOffset(consumerId string, offset int64)
	ObserveConsumerLag(consumerId string, lag int64)
	ObserveConsumerQueueLength(consumerId string, queueLength int64)
	ObserveConsumerQueueCapacity(consumerId string, queueCapacity int64)
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
		m.storage = stats.NewConsumerStorage(metrics.DefaultRegistry)
	}

	return m
}

func (m *Metrics) IsSend() bool {
	return m.isSend
}

func (m *Metrics) Send(stats kafka.ReaderStats) {
	m.storage.ObserveConsumerDials(stats.ClientID, stats.Dials)
	m.storage.ObserveConsumerFetches(stats.ClientID, stats.Fetches)
	m.storage.ObserveConsumerMessages(stats.ClientID, stats.Messages)
	m.storage.ObserveConsumerMessageBytes(stats.ClientID, stats.Bytes)
	m.storage.ObserveConsumerRebalances(stats.ClientID, stats.Rebalances)
	m.storage.ObserveConsumerTimeouts(stats.ClientID, stats.Timeouts)
	m.storage.ObserveConsumerError(stats.ClientID, stats.Errors)

	m.storage.ObserveConsumerDialTime(stats.ClientID, stats.DialTime)
	m.storage.ObserveConsumerReadTime(stats.ClientID, stats.ReadTime)
	m.storage.ObserveConsumerWaitTime(stats.ClientID, stats.WaitTime)
	m.storage.ObserveConsumerFetchSize(stats.ClientID, stats.FetchSize)
	m.storage.ObserveConsumerFetchBytes(stats.ClientID, stats.FetchBytes)

	m.storage.ObserveConsumerOffset(stats.ClientID, stats.Offset)
	m.storage.ObserveConsumerLag(stats.ClientID, stats.Lag)
	m.storage.ObserveConsumerQueueLength(stats.ClientID, stats.QueueLength)
	m.storage.ObserveConsumerQueueCapacity(stats.ClientID, stats.QueueCapacity)
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

package consumer

import (
	"github.com/segmentio/kafka-go"
)

type MetricStorage interface {
	ObserveConsumerDials(consumerId, topic string, dials int64)
	ObserveConsumerFetches(consumerId, topic string, fetches int64)
	ObserveConsumerMessages(consumerId, topic string, messages int64)
	ObserveConsumerMessageBytes(consumerId, topic string, messageBytes int64)
	ObserveConsumerRebalances(consumerId, topic string, rebalances int64)
	ObserveConsumerTimeouts(consumerId, topic string, timeouts int64)
	ObserveConsumerError(consumerId, topic string, errors int64)

	ObserveConsumerDialTime(consumerId, topic string, dialTime kafka.DurationStats)
	ObserveConsumerReadTime(consumerId, topic string, readTime kafka.DurationStats)
	ObserveConsumerWaitTime(consumerId, topic string, waitTime kafka.DurationStats)
	ObserveConsumerFetchSize(consumerId, topic string, fetchSize kafka.SummaryStats)
	ObserveConsumerFetchBytes(consumerId, topic string, fetchBytes kafka.SummaryStats)

	ObserveConsumerOffset(consumerId, topic string, offset int64)
	ObserveConsumerLag(consumerId, topic string, lag int64)
	ObserveConsumerQueueLength(consumerId, topic string, queueLength int64)
	ObserveConsumerQueueCapacity(consumerId, topic string, queueCapacity int64)
}

func (c *Consumer) sendMetrics(stats kafka.ReaderStats) {
	c.metricStorage.ObserveConsumerDials(stats.ClientID, stats.Topic, stats.Dials)
	c.metricStorage.ObserveConsumerFetches(stats.ClientID, stats.Topic, stats.Fetches)
	c.metricStorage.ObserveConsumerMessages(stats.ClientID, stats.Topic, stats.Messages)
	c.metricStorage.ObserveConsumerMessageBytes(stats.ClientID, stats.Topic, stats.Bytes)
	c.metricStorage.ObserveConsumerRebalances(stats.ClientID, stats.Topic, stats.Rebalances)
	c.metricStorage.ObserveConsumerTimeouts(stats.ClientID, stats.Topic, stats.Timeouts)
	c.metricStorage.ObserveConsumerError(stats.ClientID, stats.Topic, stats.Errors)

	c.metricStorage.ObserveConsumerDialTime(stats.ClientID, stats.Topic, stats.DialTime)
	c.metricStorage.ObserveConsumerReadTime(stats.ClientID, stats.Topic, stats.ReadTime)
	c.metricStorage.ObserveConsumerWaitTime(stats.ClientID, stats.Topic, stats.WaitTime)
	c.metricStorage.ObserveConsumerFetchSize(stats.ClientID, stats.Topic, stats.FetchSize)
	c.metricStorage.ObserveConsumerFetchBytes(stats.ClientID, stats.Topic, stats.FetchBytes)

	c.metricStorage.ObserveConsumerOffset(stats.ClientID, stats.Topic, stats.Offset)
	c.metricStorage.ObserveConsumerLag(stats.ClientID, stats.Topic, stats.Lag)
	c.metricStorage.ObserveConsumerQueueLength(stats.ClientID, stats.Topic, stats.QueueLength)
	c.metricStorage.ObserveConsumerQueueCapacity(stats.ClientID, stats.Topic, stats.QueueCapacity)
}

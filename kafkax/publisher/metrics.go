package publisher

import (
	"github.com/segmentio/kafka-go"
)

type MetricStorage interface {
	ObservePublisherWrites(publisherId, topic string, writes int64)
	ObservePublisherMessages(publisherId, topic string, messages int64)
	ObservePublisherMessageBytes(publisherId, topic string, messageBytes int64)
	ObservePublisherErrors(publisherId, topic string, errors int64)
	ObservePublisherRetries(publisherId, topic string, retries int64)

	ObserveConsumerBatchTime(publisherId, topic string, batchTime kafka.DurationStats)
	ObserveConsumerBatchQueueTime(publisherId, topic string, batchQueueTime kafka.DurationStats)
	ObserveConsumerWriteTime(publisherId, topic string, writeTime kafka.DurationStats)
	ObserveConsumerWaitTime(publisherId, topic string, waitTime kafka.DurationStats)
	ObserveConsumerBatchSize(publisherId, topic string, batchSize kafka.SummaryStats)
	ObserveConsumerBatchBytes(publisherId, topic string, batchBytes kafka.SummaryStats)
}

func (p *Publisher) sendMetrics(stats kafka.WriterStats) {
	p.metricStorage.ObservePublisherWrites(stats.ClientID, stats.Topic, stats.Writes)
	p.metricStorage.ObservePublisherMessages(stats.ClientID, stats.Topic, stats.Messages)
	p.metricStorage.ObservePublisherMessageBytes(stats.ClientID, stats.Topic, stats.Bytes)
	p.metricStorage.ObservePublisherErrors(stats.ClientID, stats.Topic, stats.Errors)
	p.metricStorage.ObservePublisherRetries(stats.ClientID, stats.Topic, stats.Retries)

	p.metricStorage.ObserveConsumerBatchTime(stats.ClientID, stats.Topic, stats.BatchTime)
	p.metricStorage.ObserveConsumerBatchQueueTime(stats.ClientID, stats.Topic, stats.BatchQueueTime)
	p.metricStorage.ObserveConsumerWriteTime(stats.ClientID, stats.Topic, stats.WriteTime)
	p.metricStorage.ObserveConsumerWaitTime(stats.ClientID, stats.Topic, stats.WaitTime)
	p.metricStorage.ObserveConsumerBatchSize(stats.ClientID, stats.Topic, stats.BatchSize)
	p.metricStorage.ObserveConsumerBatchBytes(stats.ClientID, stats.Topic, stats.BatchBytes)
}

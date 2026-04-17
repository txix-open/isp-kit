package kafka_metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

// ConsumerStorage collects metrics for Kafka consumer operations, including message
// consume latency, message body sizes, commit counts, and retry counts.
type ConsumerStorage struct {
	consumeMsgDuration *prometheus.SummaryVec
	consumeMsgBodySize *prometheus.SummaryVec
	commitCount        *prometheus.CounterVec
	retryCount         *prometheus.CounterVec
}

// NewConsumerStorage creates a new ConsumerStorage instance and registers its metrics
// with the provided registry. Metrics are labeled by consumer group and topic.
func NewConsumerStorage(reg *metrics.Registry) *ConsumerStorage {
	s := &ConsumerStorage{
		consumeMsgDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "kafka",
			Name:       "consume_duration_ms",
			Help:       "The latency of handling single message from topic",
			Objectives: metrics.DefaultObjectives,
		}, []string{"consumerGroup", "topic"})),
		consumeMsgBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "kafka",
			Name:       "consume_body_size",
			Help:       "The size of message body from queue",
			Objectives: metrics.DefaultObjectives,
		}, []string{"consumerGroup", "topic"})),
		commitCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "kafka",
			Name:      "consume_commit_count",
			Help:      "Count of committed messages",
		}, []string{"consumerGroup", "topic"})),
		retryCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "kafka",
			Name:      "consume_retry_count",
			Help:      "Count of retried messages",
		}, []string{"consumerGroup", "topic"})),
	}
	return s
}

// ObserveConsumeDuration records the latency of processing a single message.
func (c *ConsumerStorage) ObserveConsumeDuration(consumerGroup, topic string, t time.Duration) {
	c.consumeMsgDuration.WithLabelValues(consumerGroup, topic).Observe(metrics.Milliseconds(t))
}

// ObserveConsumeMsgSize records the size of a consumed message in bytes.
func (c *ConsumerStorage) ObserveConsumeMsgSize(consumerGroup, topic string, size int) {
	c.consumeMsgBodySize.WithLabelValues(consumerGroup, topic).Observe(float64(size))
}

// IncCommitCount increments the counter for successfully committed messages.
func (c *ConsumerStorage) IncCommitCount(consumerGroup, topic string) {
	c.commitCount.WithLabelValues(consumerGroup, topic).Inc()
}

// IncRetryCount increments the counter for retried message processing.
func (c *ConsumerStorage) IncRetryCount(consumerGroup, topic string) {
	c.retryCount.WithLabelValues(consumerGroup, topic).Inc()
}

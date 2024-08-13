package kafka_metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

type ConsumerStorage struct {
	consumeMsgDuration *prometheus.SummaryVec
	consumeMsgBodySize *prometheus.SummaryVec
	commitCount        *prometheus.CounterVec
	retryCount         *prometheus.CounterVec
}

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
			Help:      "Count of commited messages",
		}, []string{"consumerGroup", "topic"})),
		retryCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "kafka",
			Name:      "consume_retry_count",
			Help:      "Count of retried messages",
		}, []string{"consumerGroup", "topic"})),
	}
	return s
}

func (c *ConsumerStorage) ObserveConsumeDuration(consumerGroup, topic string, t time.Duration) {
	c.consumeMsgDuration.WithLabelValues(consumerGroup, topic).Observe(metrics.Milliseconds(t))
}

func (c *ConsumerStorage) ObserveConsumeMsgSize(consumerGroup, topic string, size int) {
	c.consumeMsgBodySize.WithLabelValues(consumerGroup, topic).Observe(float64(size))
}

func (c *ConsumerStorage) IncCommitCount(consumerGroup, topic string) {
	c.commitCount.WithLabelValues(consumerGroup, topic).Inc()
}

func (c *ConsumerStorage) IncRetryCount(consumerGroup, topic string) {
	c.retryCount.WithLabelValues(consumerGroup, topic).Inc()
}

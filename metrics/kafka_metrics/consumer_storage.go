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
		}, []string{"topic"})),
		consumeMsgBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "kafka",
			Name:       "consume_body_size",
			Help:       "The size of message body from queue",
			Objectives: metrics.DefaultObjectives,
		}, []string{"topic"})),
		commitCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "kafka",
			Name:      "consume_commit_count",
			Help:      "Count of commited messages",
		}, []string{"topic"})),
		retryCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "kafka",
			Name:      "consume_retry_count",
			Help:      "Count of retried messages",
		}, []string{"topic"})),
	}
	return s
}

func (c *ConsumerStorage) ObserveConsumeDuration(topic string, t time.Duration) {
	c.consumeMsgDuration.WithLabelValues(topic).Observe(metrics.Milliseconds(t))
}

func (c *ConsumerStorage) ObserveConsumeMsgSize(topic string, size int) {
	c.consumeMsgBodySize.WithLabelValues(topic).Observe(float64(size))
}

func (c *ConsumerStorage) IncCommitCount(topic string) {
	c.commitCount.WithLabelValues(topic).Inc()
}

func (c *ConsumerStorage) IncRetryCount(topic string) {
	c.retryCount.WithLabelValues(topic).Inc()
}

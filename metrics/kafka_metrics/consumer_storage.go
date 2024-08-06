package kafka_metrics

import (
	"strconv"
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
		}, []string{"topic", "partition", "offset"})),
		consumeMsgBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "kafka",
			Name:       "consume_body_size",
			Help:       "The size of message body from queue",
			Objectives: metrics.DefaultObjectives,
		}, []string{"topic", "partition", "offset"})),
		commitCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "kafka",
			Name:      "consume_commit_count",
			Help:      "Count of commited messages",
		}, []string{"topic", "partition", "offset"})),
		retryCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "kafka",
			Name:      "consume_retry_count",
			Help:      "Count of retried messages",
		}, []string{"topic", "partition", "offset"})),
	}
	return s
}

func (c *ConsumerStorage) ObserveConsumeDuration(topic string, partition int, offset int64, t time.Duration) {
	c.consumeMsgDuration.WithLabelValues(topic, strconv.Itoa(partition), strconv.Itoa(int(offset))).
		Observe(metrics.Milliseconds(t))
}

func (c *ConsumerStorage) ObserveConsumeMsgSize(topic string, partition int, offset int64, size int) {
	c.consumeMsgBodySize.WithLabelValues(topic, strconv.Itoa(partition), strconv.Itoa(int(offset))).
		Observe(float64(size))
}

func (c *ConsumerStorage) IncCommitCount(topic string, partition int, offset int64) {
	c.commitCount.WithLabelValues(topic, strconv.Itoa(partition), strconv.Itoa(int(offset))).Inc()
}

func (c *ConsumerStorage) IncRetryCount(topic string, partition int, offset int64) {
	c.retryCount.WithLabelValues(topic, strconv.Itoa(partition), strconv.Itoa(int(offset))).Inc()
}

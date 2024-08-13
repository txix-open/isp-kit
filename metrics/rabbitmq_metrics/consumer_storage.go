package rabbitmq_metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

type ConsumerStorage struct {
	consumeMsgDuration *prometheus.SummaryVec
	consumeMsgBodySize *prometheus.SummaryVec
	dlqCount           *prometheus.CounterVec
	requeueCount       *prometheus.CounterVec
	retryCount         *prometheus.CounterVec
	successCount       *prometheus.CounterVec
}

func NewConsumerStorage(reg *metrics.Registry) *ConsumerStorage {
	s := &ConsumerStorage{
		consumeMsgDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "rabbitmq",
			Name:       "consume_duration_ms",
			Help:       "The latency of handling single message from queue",
			Objectives: metrics.DefaultObjectives,
		}, []string{"exchange", "routing_key"})),
		consumeMsgBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "rabbitmq",
			Name:       "consume_body_size",
			Help:       "The size of message body from queue",
			Objectives: metrics.DefaultObjectives,
		}, []string{"exchange", "routing_key"})),
		requeueCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "rabbitmq",
			Name:      "consume_requeue_count",
			Help:      "Count of requeued messages",
		}, []string{"exchange", "routing_key"})),
		dlqCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "rabbitmq",
			Name:      "consume_dlq_count",
			Help:      "Count of messages moved to DLQ",
		}, []string{"exchange", "routing_key"})),
		successCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "rabbitmq",
			Name:      "consume_success_count",
			Help:      "Count of successful messages",
		}, []string{"exchange", "routing_key"})),
		retryCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "rabbitmq",
			Name:      "consume_retry_count",
			Help:      "Count of retried messages",
		}, []string{"exchange", "routing_key"})),
	}
	return s
}

func (c *ConsumerStorage) ObserveConsumeDuration(exchange string, routingKey string, duration time.Duration) {
	c.consumeMsgDuration.WithLabelValues(exchange, routingKey).Observe(metrics.Milliseconds(duration))
}

func (c *ConsumerStorage) ObserveConsumeMsgSize(exchange string, routingKey string, size int) {
	c.consumeMsgBodySize.WithLabelValues(exchange, routingKey).Observe(float64(size))
}

func (c *ConsumerStorage) IncRequeueCount(exchange string, routingKey string) {
	c.requeueCount.WithLabelValues(exchange, routingKey).Inc()
}

func (c *ConsumerStorage) IncDlqCount(exchange string, routingKey string) {
	c.dlqCount.WithLabelValues(exchange, routingKey).Inc()
}

func (c *ConsumerStorage) IncSuccessCount(exchange string, routingKey string) {
	c.successCount.WithLabelValues(exchange, routingKey).Inc()
}

func (c *ConsumerStorage) IncRetryCount(exchange string, routingKey string) {
	c.retryCount.WithLabelValues(exchange, routingKey).Inc()
}

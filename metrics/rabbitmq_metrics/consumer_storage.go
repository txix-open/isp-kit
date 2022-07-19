package rabbitmq_metircs

import (
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type ConsumerStorage struct {
	consumeMsgDuration *prometheus.HistogramVec
	consumeMsgBodySize *prometheus.HistogramVec
	dlqCount           *prometheus.CounterVec
	requeueCount       *prometheus.CounterVec
	successCount       *prometheus.CounterVec
}

func NewConsumerStorage(reg *metrics.Registry) *ConsumerStorage {
	s := &ConsumerStorage{
		consumeMsgDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "rabbitmq",
			Name:      "consume_duration_ms",
			Help:      "The latency of handling single message from queue",
			Buckets:   metrics.DefaultDurationMsBuckets,
		}, []string{"exchange", "routing_key"}),
		consumeMsgBodySize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "rabbitmq",
			Name:      "consume_body_size",
			Help:      "The size of message body from queue",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		}, []string{"exchange", "routing_key"}),
		requeueCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "rabbitmq",
			Name:      "consume_requeue_count",
			Help:      "Count of requeued messages",
		}, []string{"exchange", "routing_key"}),
		dlqCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "rabbitmq",
			Name:      "consume_dlq_count",
			Help:      "Count of messages moved to DLQ",
		}, []string{"exchange", "routing_key"}),
		successCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "rabbitmq",
			Name:      "consume_success_count",
			Help:      "Count of successful messages",
		}, []string{"exchange", "routing_key"}),
	}
	s.consumeMsgDuration = reg.GetOrRegister(s.consumeMsgDuration).(*prometheus.HistogramVec)
	s.consumeMsgBodySize = reg.GetOrRegister(s.consumeMsgBodySize).(*prometheus.HistogramVec)
	s.requeueCount = reg.GetOrRegister(s.requeueCount).(*prometheus.CounterVec)
	s.dlqCount = reg.GetOrRegister(s.dlqCount).(*prometheus.CounterVec)
	s.successCount = reg.GetOrRegister(s.successCount).(*prometheus.CounterVec)
	return s
}

func (c *ConsumerStorage) ObserveConsumeDuration(exchange string, routingKey string, t time.Duration) {
	c.consumeMsgDuration.WithLabelValues(exchange, routingKey).Observe(float64(t.Milliseconds()))
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

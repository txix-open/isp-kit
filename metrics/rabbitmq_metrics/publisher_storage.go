package rabbitmq_metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

// PublisherStorage collects metrics for RabbitMQ publisher operations, including message
// publish latency, message body sizes, and publish error counts.
type PublisherStorage struct {
	publishMsgDuration *prometheus.SummaryVec
	publishMsgBodySize *prometheus.SummaryVec
	publishErrorCount  *prometheus.CounterVec
}

// NewPublisherStorage creates a new PublisherStorage instance and registers its metrics
// with the provided registry. Metrics are labeled by exchange and routing key.
func NewPublisherStorage(reg *metrics.Registry) *PublisherStorage {
	s := &PublisherStorage{
		publishMsgDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "rabbitmq",
			Name:       "publish_duration_ms",
			Help:       "The latency of publishing single message to queue",
			Objectives: metrics.DefaultObjectives,
		}, []string{"exchange", "routing_key"})),
		publishMsgBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "rabbitmq",
			Name:       "publish_body_size",
			Help:       "The size of published message body to queue",
			Objectives: metrics.DefaultObjectives,
		}, []string{"exchange", "routing_key"})),
		publishErrorCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "rabbitmq",
			Name:      "publish_error_count",
			Help:      "Count error on publishing",
		}, []string{"exchange", "routing_key"})),
	}
	return s
}

// ObservePublishDuration records the latency of publishing a message to RabbitMQ.
func (c *PublisherStorage) ObservePublishDuration(exchange string, routingKey string, duration time.Duration) {
	c.publishMsgDuration.WithLabelValues(exchange, routingKey).Observe(metrics.Milliseconds(duration))
}

// ObservePublishMsgSize records the size of a published message in bytes.
func (c *PublisherStorage) ObservePublishMsgSize(exchange string, routingKey string, size int) {
	c.publishMsgBodySize.WithLabelValues(exchange, routingKey).Observe(float64(size))
}

// IncPublishError increments the error counter for publishing failures.
func (c *PublisherStorage) IncPublishError(exchange string, routingKey string) {
	c.publishErrorCount.WithLabelValues(exchange, routingKey).Inc()
}

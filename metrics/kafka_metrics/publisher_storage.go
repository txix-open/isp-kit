package kafka_metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

// PublisherStorage collects metrics for Kafka producer operations, including message
// publish latency, message body sizes, and publish error counts.
type PublisherStorage struct {
	publishMsgDuration *prometheus.SummaryVec
	publishMsgBodySize *prometheus.SummaryVec
	publishErrorCount  *prometheus.CounterVec
}

// NewPublisherStorage creates a new PublisherStorage instance and registers its metrics
// with the provided registry. Metrics are labeled by Kafka topic.
func NewPublisherStorage(reg *metrics.Registry) *PublisherStorage {
	s := &PublisherStorage{
		publishMsgDuration: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "kafka",
			Name:       "publish_duration_ms",
			Help:       "The latency of publishing messages to topic",
			Objectives: metrics.DefaultObjectives,
		}, []string{"topic"})),
		publishMsgBodySize: metrics.GetOrRegister(reg, prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Subsystem:  "kafka",
			Name:       "publish_body_size",
			Help:       "The size of published message body to topic",
			Objectives: metrics.DefaultObjectives,
		}, []string{"topic"})),
		publishErrorCount: metrics.GetOrRegister(reg, prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "kafka",
			Name:      "publish_error_count",
			Help:      "Count error on publishing",
		}, []string{"topic"})),
	}
	return s
}

// ObservePublishDuration records the latency of publishing a message to a Kafka topic.
func (c *PublisherStorage) ObservePublishDuration(topic string, t time.Duration) {
	c.publishMsgDuration.WithLabelValues(topic).Observe(metrics.Milliseconds(t))
}

// ObservePublishMsgSize records the size of a published message in bytes.
func (c *PublisherStorage) ObservePublishMsgSize(topic string, size int) {
	c.publishMsgBodySize.WithLabelValues(topic).Observe(float64(size))
}

// IncPublishError increments the error counter for publishing failures to a topic.
func (c *PublisherStorage) IncPublishError(topic string) {
	c.publishErrorCount.WithLabelValues(topic).Inc()
}

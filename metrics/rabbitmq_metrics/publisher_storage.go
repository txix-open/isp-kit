package rabbitmq_metircs

import (
	"time"

	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type PublisherStorage struct {
	publishMsgDuration *prometheus.HistogramVec
	publishMsgBodySize *prometheus.HistogramVec
	publishErrorCount  *prometheus.CounterVec
}

func NewPublisherStorage(reg *metrics.Registry) *PublisherStorage {
	s := &PublisherStorage{
		publishMsgDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "rabbitmq",
			Name:      "publish_duration_ms",
			Help:      "The latency of publishing single message to queue",
			Buckets:   metrics.DefaultDurationMsBuckets,
		}, []string{"exchange", "routing_key"}),
		publishMsgBodySize: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "rabbitmq",
			Name:      "publish_body_size",
			Help:      "The size of published message body to queue",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8),
		}, []string{"exchange", "routing_key"}),
		publishErrorCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "rabbitmq",
			Name:      "publish_error_count",
			Help:      "Count error on publishing",
		}, []string{"exchange", "routing_key"}),
	}
	s.publishMsgDuration = reg.GetOrRegister(s.publishMsgDuration).(*prometheus.HistogramVec)
	s.publishMsgBodySize = reg.GetOrRegister(s.publishMsgBodySize).(*prometheus.HistogramVec)
	s.publishErrorCount = reg.GetOrRegister(s.publishErrorCount).(*prometheus.CounterVec)
	return s
}

func (c *PublisherStorage) ObservePublishDuration(exchange string, routingKey string, t time.Duration) {
	c.publishMsgDuration.WithLabelValues(exchange, routingKey).Observe(float64(t.Milliseconds()))
}

func (c *PublisherStorage) ObservePublishMsgSize(exchange string, routingKey string, size int) {
	c.publishMsgBodySize.WithLabelValues(exchange, routingKey).Observe(float64(size))
}

func (c *PublisherStorage) IncPublishError(exchange string, routingKey string) {
	c.publishErrorCount.WithLabelValues(exchange, routingKey).Inc()
}

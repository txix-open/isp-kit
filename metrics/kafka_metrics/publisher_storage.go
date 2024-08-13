package kafka_metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/metrics"
)

type PublisherStorage struct {
	publishMsgDuration *prometheus.SummaryVec
	publishMsgBodySize *prometheus.SummaryVec
	publishErrorCount  *prometheus.CounterVec
}

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

func (c *PublisherStorage) ObservePublishDuration(topic string, t time.Duration) {
	c.publishMsgDuration.WithLabelValues(topic).Observe(metrics.Milliseconds(t))
}

func (c *PublisherStorage) ObservePublishMsgSize(topic string, size int) {
	c.publishMsgBodySize.WithLabelValues(topic).Observe(float64(size))
}

func (c *PublisherStorage) IncPublishError(topic string) {
	c.publishErrorCount.WithLabelValues(topic).Inc()
}

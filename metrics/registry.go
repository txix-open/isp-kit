// Package metrics provides a Prometheus-based metrics collection system with support for
// various subsystems including HTTP, gRPC, Kafka, RabbitMQ, SQL, and background jobs.
//
// The package offers a thread-safe registry for managing Prometheus collectors, along with
// pre-configured storage types for common integration patterns.
//
// Example usage:
//
//	reg := metrics.NewRegistry()
//	httpStorage := http_metrics.NewServerStorage(reg)
//	http.HandleFunc("/metrics", reg.MetricsHandler())
package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

const (
	describeChanCapacity      = 512
	metricsDescriptionTimeout = 1 * time.Second
)

// Metric is an alias for Prometheus collector interface.
type Metric interface {
	prometheus.Collector
}

// Registry provides a thread-safe wrapper around Prometheus registry for managing
// metric collectors. It automatically registers default Go, build info, and process
// collectors on creation.
type Registry struct {
	lock sync.Locker
	reg  *prometheus.Registry
	list []Metric
}

// NewRegistry creates a new metric registry with default collectors for Go runtime,
// build information, and process statistics.
func NewRegistry() *Registry {
	r := &Registry{
		reg:  prometheus.NewRegistry(),
		lock: &sync.Mutex{},
	}
	r.GetOrRegister(collectors.NewGoCollector())
	r.GetOrRegister(collectors.NewBuildInfoCollector())
	r.GetOrRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	return r
}

// GetOrRegister attempts to register a metric collector. If the metric is already
// registered, it returns the existing collector instead. Panics if registration fails
// for any other reason.
func (r *Registry) GetOrRegister(metric Metric) Metric {
	r.lock.Lock()
	defer r.lock.Unlock()

	err := r.reg.Register(metric)
	alreadyExist := prometheus.AlreadyRegisteredError{}
	if errors.As(err, &alreadyExist) {
		return alreadyExist.ExistingCollector
	}
	if err != nil {
		panic(errors.WithMessagef(err, "metrics registry: register %v", metric))
	}
	r.list = append(r.list, metric)
	return metric
}

// MetricsHandler returns an HTTP handler that exposes all registered metrics in
// Prometheus text format. This handler should be mounted on an endpoint (typically
// "/metrics") for scraping by Prometheus.
func (r *Registry) MetricsHandler() http.Handler {
	handler := promhttp.InstrumentMetricHandler(
		r.reg, promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{}),
	)
	return handler
}

// MetricsDescriptionHandler returns an HTTP handler that provides human-readable
// descriptions of all registered metrics. The handler is limited to a 1 second
// timeout to prevent blocking.
func (r *Registry) MetricsDescriptionHandler() http.Handler {
	return http.TimeoutHandler(
		http.HandlerFunc(r.metricsDescriptionHandler),
		metricsDescriptionTimeout,
		"timeout",
	)
}

// Register adds a collector to the registry. This is a convenience method that
// calls GetOrRegister and always returns nil.
func (r *Registry) Register(collector prometheus.Collector) error {
	r.GetOrRegister(collector)
	return nil
}

// MustRegister registers multiple collectors. Panics if any collector fails to
// register.
func (r *Registry) MustRegister(collectors ...prometheus.Collector) {
	for _, c := range collectors {
		r.GetOrRegister(c)
	}
}

// Unregister removes a collector from the registry. Returns true if the collector
// was successfully unregistered, false otherwise.
func (r *Registry) Unregister(collector prometheus.Collector) bool {
	return r.reg.Unregister(collector)
}

// Gather collects all registered metrics and returns them as protobuf-encoded
// metric families. This method can be used to programmatically access metric data.
func (r *Registry) Gather() ([]*io_prometheus_client.MetricFamily, error) {
	return r.reg.Gather()
}

func (r *Registry) metricsDescriptionHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")

	for _, metric := range r.list {
		c := make(chan *prometheus.Desc, describeChanCapacity)
		metric.Describe(c)
		for range len(c) {
			desc := <-c
			_, _ = fmt.Fprintf(writer, "%s, type: %T\n", desc, metric)
		}
	}
}

// GetOrRegister is a generic wrapper around Registry.GetOrRegister that preserves
// the concrete type of the metric.
func GetOrRegister[M Metric](registry *Registry, metric M) M {
	m := registry.GetOrRegister(metric)
	return m.(M) // nolint:forcetypeassert
}

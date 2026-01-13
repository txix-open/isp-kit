// nolint:ireturn
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

type Metric interface {
	prometheus.Collector
}

type Registry struct {
	lock sync.Locker
	reg  *prometheus.Registry
	list []Metric
}

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

func (r *Registry) MetricsHandler() http.Handler {
	handler := promhttp.InstrumentMetricHandler(
		r.reg, promhttp.HandlerFor(r.reg, promhttp.HandlerOpts{}),
	)
	return handler
}

func (r *Registry) MetricsDescriptionHandler() http.Handler {
	return http.TimeoutHandler(
		http.HandlerFunc(r.metricsDescriptionHandler),
		metricsDescriptionTimeout,
		"timeout",
	)
}

func (r *Registry) Register(collector prometheus.Collector) error {
	r.GetOrRegister(collector)
	return nil
}

func (r *Registry) MustRegister(collectors ...prometheus.Collector) {
	for _, c := range collectors {
		r.GetOrRegister(c)
	}
}

func (r *Registry) Unregister(collector prometheus.Collector) bool {
	return r.reg.Unregister(collector)
}

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

func GetOrRegister[M Metric](registry *Registry, metric M) M {
	m := registry.GetOrRegister(metric)
	return m.(M) // nolint:forcetypeassert
}

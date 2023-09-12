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
	return http.TimeoutHandler(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("content-type", "text/plain")

		for _, metric := range r.list {
			c := make(chan *prometheus.Desc, 512)
			metric.Describe(c)
			for i := 0; i < len(c); i++ {
				desc := <-c
				_, _ = fmt.Fprintf(writer, "%s, type: %T\n", desc, metric)
			}
		}
	}), 1*time.Second, "timeout")
}

func GetOrRegister[M Metric](registry *Registry, metric M) M {
	m := registry.GetOrRegister(metric)
	return m.(M)
}

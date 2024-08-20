package mcounter

import (
	"context"
	"crypto/sha256"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"slices"
	"sync"
	"time"
)

var (
	DuplicateNameErr = errors.New("duplicate name")
)

type MetricRep interface {
	Counters() ([]Counter, error)
	CounterValues(counterId string) ([]CounterValue, error)

	UpsertCounterBatched(ctx context.Context, counter []*Counter) error
	UpsertCounterValueBatched(ctx context.Context, counter []*CounterValue) error
}

type Counter struct {
	Name   string
	Labels []string

	counterValue map[string]*CounterValue
}

type CounterValue struct {
	// id = hash(CounterName + LabelValues)
	id string

	CounterName string
	LabelValues []string
	AddValue    int
}

type Metrics struct {
	log   log.Logger
	close chan struct{}
	// ctx global
	ctx context.Context

	buffMu        sync.RWMutex
	counterBuffer map[string]*Counter
	bufferCap     uint
	// bufferLen overall amount of CounterValue-s
	// inside Counter.counterValue across all Counter-s
	bufferLen uint

	registry   *metrics.Registry
	metricsRep MetricRep
}

func NewMetrics(ctx context.Context, log log.Logger, metricsRep MetricRep, bufferCap uint, flushInterval time.Duration) (*Metrics, error) {
	metrics := &Metrics{
		counterBuffer: make(map[string]*Counter),
		log:           log,
		ctx:           ctx,
		close:         make(chan struct{}),

		bufferCap:  bufferCap,
		metricsRep: metricsRep,
	}

	metrics.runTimedFlusher(ctx, flushInterval)
	err := metrics.load()
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (m *Metrics) Inc(name string, meta map[string]string) error {
	var (
		labelNames      = make([]string, len(meta))
		labelValuesHash = sha256.New()

		counter      *Counter
		counterValue *CounterValue
		exist        bool
	)

	labelValuesHash.Write([]byte(name))
	for key, value := range meta {
		labelNames = append(labelNames, key)
		labelValuesHash.Write([]byte(value))
	}
	reg := m.register(name, labelNames)
	if reg == nil {
		return DuplicateNameErr
	}

	reg.With(meta).Inc()

	m.buffMu.Lock()
	defer m.buffMu.Unlock()

	if counter, exist = m.counterBuffer[name]; !exist {
		counter = &Counter{
			Name:         counter.Name,
			Labels:       labelNames,
			counterValue: make(map[string]*CounterValue),
		}
		m.counterBuffer[name] = counter
	}
	labelValuesHashStr := string(labelValuesHash.Sum(nil))
	if counterValue, exist = counter.counterValue[labelValuesHashStr]; !exist {
		counterValue = &CounterValue{
			id:          labelValuesHashStr,
			CounterName: name,
		}
		counter.counterValue[labelValuesHashStr] = counterValue
		m.bufferLen++
	}
	counterValue.AddValue++

	if m.bufferLen > m.bufferCap {
		go func() {
			if err := m.flush(m.ctx); err == nil {
				return
			}
			m.log.Error(
				context.Background(),
				"metrics: saving counter metrics to db",
			)
		}()
	}

	return nil
}

func (m *Metrics) Close(ctx context.Context) error {
	var err error
	if err = m.flush(ctx); err == nil {
		return nil
	}
	m.close <- struct{}{}

	return err
}

func (m *Metrics) load() error {
	counters, err := m.metricsRep.Counters()
	if err != nil {
		return err
	}
	for _, counter := range counters {
		counterValues, err := m.metricsRep.CounterValues(counter.Name)
		if err != nil {
			return err
		}
		reg := m.register(counter.Name, counter.Labels)
		if reg == nil {
			return DuplicateNameErr
		}
		for _, val := range counterValues {
			reg.WithLabelValues(val.LabelValues...).Add(float64(val.AddValue))
		}
	}

	return nil
}

func (m *Metrics) register(name string, labelNames []string) *prometheus.CounterVec {
	slices.Sort(labelNames)
	counterMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
		}, labelNames)

	defer func() {
		recover()
	}()
	reg := metrics.GetOrRegister(m.registry, counterMetric)

	return reg
}

func (m *Metrics) flush(ctx context.Context) error {
	var (
		counters      = make([]*Counter, len(m.counterBuffer))
		counterValues = make([]*CounterValue, m.bufferCap)
	)
	m.buffMu.RLock()
	defer m.buffMu.RUnlock()

	for _, counter := range m.counterBuffer {
		counters = append(counters, counter)
		for _, counterValue := range counter.counterValue {
			counterValues = append(counterValues, counterValue)
		}
	}
	err := m.metricsRep.UpsertCounterBatched(ctx, counters)
	if err != nil {
		return err
	}
	err = m.metricsRep.UpsertCounterValueBatched(ctx, counterValues)
	if err != nil {
		return err
	}

	m.counterBuffer = make(map[string]*Counter)
	m.bufferLen = 0

	return nil
}

func (m *Metrics) runTimedFlusher(ctx context.Context, interval time.Duration) {
	if interval == 0 {
		return
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.close:
			return
		case <-timer.C:
			err := m.flush(ctx)
			if err != nil {
				m.log.Error(ctx, "metrics: saving counter metrics to db")
				m.close <- struct{}{}
				return
			}
		}
	}
}

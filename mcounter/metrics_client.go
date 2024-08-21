package mcounter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"github.com/lib/pq"
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

type CounterTransaction interface {
	UpsertCounter(ctx context.Context, counter []*counter) error
	UpsertCounterValue(ctx context.Context, counter []*counterValue) error
}

type CounterTransactionRunner interface {
	CounterTransaction(ctx context.Context, tx func(ctx context.Context, tx CounterTransaction) error) error
}

type MetricRep interface {
	Counters(ctx context.Context) ([]counter, error)
	CounterValues(ctx context.Context, counterId string) ([]counterValue, error)

	UpsertCounter(ctx context.Context, counter []*counter) error
	UpsertCounterValue(ctx context.Context, counter []*counterValue) error
}

type counter struct {
	Name   string         `db:"name"`
	Labels pq.StringArray `db:"labels"`

	counterValues map[string]*counterValue
}

type counterValue struct {
	// id = hash(CounterName + LabelValues)
	Id          string         `db:"id"`
	CounterName string         `db:"counter_name"`
	LabelValues pq.StringArray `db:"label_values"`
	AddValue    int            `db:"counter_value"`
}

type CounterConfig struct {
	BufferCap     uint
	FlushInterval time.Duration
}

func DefaultConfig() *CounterConfig {
	return &CounterConfig{
		BufferCap:     10,
		FlushInterval: 5 * time.Second,
	}
}

func (c *CounterConfig) WithFlushInterval(interval time.Duration) *CounterConfig {
	c.FlushInterval = interval
	return c
}

func (c *CounterConfig) WithBufferCap(capacity uint) *CounterConfig {
	c.BufferCap = capacity
	return c
}

type CounterMetrics struct {
	log   log.Logger
	close chan struct{}
	// ctx global
	ctx context.Context

	buffMu        sync.RWMutex
	counterBuffer map[string]*counter
	bufferCap     uint
	// bufferLen overall amount of counterValue-s
	// inside counter.counterValues across all counter-s
	bufferLen uint

	registry        *metrics.Registry
	metricsRep      MetricRep
	counterTxRunner CounterTransactionRunner
}

func NewCounterMetrics(
	ctx context.Context, registry *metrics.Registry,
	log log.Logger, metricsRep MetricRep,
	counterTxRunner CounterTransactionRunner,
	config *CounterConfig) (*CounterMetrics, error) {
	counterMetrics := &CounterMetrics{
		counterBuffer: make(map[string]*counter),
		log:           log,
		ctx:           ctx,
		close:         make(chan struct{}),

		registry:        registry,
		bufferCap:       config.BufferCap,
		metricsRep:      metricsRep,
		counterTxRunner: counterTxRunner,
	}

	err := counterMetrics.load()
	go counterMetrics.runTimedFlusher(ctx, config.FlushInterval)

	if err != nil {
		return nil, err
	}

	return counterMetrics, nil
}

func (m *CounterMetrics) Inc(name string, meta map[string]string) error {
	var (
		labelValuesHash = sha256.New()
		labelNames      []string
		labelValues     []string

		countr      *counter
		countrValue *counterValue
		exist       bool
	)

	labelValuesHash.Write([]byte(name))
	for key := range meta {
		labelNames = append(labelNames, key)
	}
	slices.Sort(labelNames)
	for _, label := range labelNames {
		labelValuesHash.Write([]byte(meta[label]))
		labelValues = append(labelValues, meta[label])
	}

	reg := m.register(name, labelNames)
	if reg == nil {
		return DuplicateNameErr
	}

	reg.With(meta).Inc()

	m.buffMu.Lock()
	defer m.buffMu.Unlock()

	if countr, exist = m.counterBuffer[name]; !exist {
		countr = &counter{
			Name:          name,
			Labels:        labelNames,
			counterValues: make(map[string]*counterValue),
		}
		m.counterBuffer[name] = countr
	}
	labelValuesHashStr := hex.EncodeToString(labelValuesHash.Sum(nil))
	if countrValue, exist = countr.counterValues[labelValuesHashStr]; !exist {
		countrValue = &counterValue{
			Id:          labelValuesHashStr,
			CounterName: name,
			LabelValues: labelValues,
		}
		countr.counterValues[labelValuesHashStr] = countrValue
		m.bufferLen++
	}
	countrValue.AddValue++

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

func (m *CounterMetrics) Close(ctx context.Context) error {
	var err error
	if err = m.flush(ctx); err == nil {
		return nil
	}
	m.close <- struct{}{}

	return err
}

func (m *CounterMetrics) load() error {
	counters, err := m.metricsRep.Counters(m.ctx)
	if err != nil {
		return err
	}
	for _, counter := range counters {
		counterValues, err := m.metricsRep.CounterValues(m.ctx, counter.Name)
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

func (m *CounterMetrics) register(name string, labelNames []string) *prometheus.CounterVec {
	slices.Sort(labelNames)
	counterMetric := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: name,
		}, labelNames)

	defer func() {
		_ = recover()
	}()
	reg := metrics.GetOrRegister(m.registry, counterMetric)

	return reg
}

func (m *CounterMetrics) flush(ctx context.Context) error {
	var (
		counters      []*counter
		counterValues []*counterValue
	)
	m.buffMu.RLock()
	defer m.buffMu.RUnlock()

	for _, counter := range m.counterBuffer {
		counters = append(counters, counter)
		for _, counterValue := range counter.counterValues {
			counterValues = append(counterValues, counterValue)
		}
	}
	err := m.counterTxRunner.CounterTransaction(ctx, func(ctx context.Context, tx CounterTransaction) error {
		return tx.UpsertCounter(ctx, counters)
	})
	if err != nil {
		return err
	}
	err = m.counterTxRunner.CounterTransaction(ctx, func(ctx context.Context, tx CounterTransaction) error {
		return tx.UpsertCounterValue(ctx, counterValues)
	})
	if err != nil {
		return err
	}

	m.counterBuffer = make(map[string]*counter)
	m.bufferLen = 0

	return nil
}

func (m *CounterMetrics) runTimedFlusher(ctx context.Context, interval time.Duration) {
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

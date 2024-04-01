package app_metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/metrics"
	"go.uber.org/zap/zapcore"
)

type keeper struct {
	sampled prometheus.Counter
	dropped prometheus.Counter
}

type LogCounter struct {
	counters map[log.Level]keeper
}

func NewLogCounter(registry *metrics.Registry) LogCounter {
	counter := metrics.GetOrRegister[*prometheus.CounterVec](
		registry,
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "app",
			Name:      "logs_sampling_status_count",
			Help:      "Count of logs statuses(dropped or sampled)",
		}, []string{"level", "status"}),
	)
	wellKnownLevels := []log.Level{
		log.FatalLevel,
		log.ErrorLevel,
		log.WarnLevel,
		log.InfoLevel,
		log.DebugLevel,
	}
	counters := make(map[log.Level]keeper)
	for _, level := range wellKnownLevels {
		counters[level] = keeper{
			sampled: counter.WithLabelValues(level.String(), "sampled"),
			dropped: counter.WithLabelValues(level.String(), "dropped"),
		}
	}
	return LogCounter{
		counters: counters,
	}
}

func (c LogCounter) SampledLogCounter() func(entry zapcore.Entry) error {
	return func(entry zapcore.Entry) error {
		keeper, ok := c.counters[entry.Level]
		if !ok {
			return nil
		}
		keeper.sampled.Inc()
		return nil
	}
}

func (c LogCounter) DroppedLogCounter() func(zapcore.Entry, zapcore.SamplingDecision) {
	return func(entry zapcore.Entry, decision zapcore.SamplingDecision) {
		if decision != zapcore.LogDropped {
			return
		}
		keeper, ok := c.counters[entry.Level]
		if !ok {
			return
		}
		keeper.dropped.Inc()
	}
}

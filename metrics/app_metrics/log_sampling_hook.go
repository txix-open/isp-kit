package app_metrics

import (
	"github.com/integration-system/isp-kit/log"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap/zapcore"
)

type keeper struct {
	sampled prometheus.Counter
	dropped prometheus.Counter
}

func LogSamplingHook(registry *metrics.Registry) func(zapcore.Entry, zapcore.SamplingDecision) {
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
	return func(entry zapcore.Entry, decision zapcore.SamplingDecision) {
		keeper, ok := counters[entry.Level]
		if !ok {
			return
		}
		switch decision {
		case zapcore.LogSampled:
			keeper.sampled.Inc()
		case zapcore.LogDropped:
			keeper.dropped.Inc()
		}
	}
}

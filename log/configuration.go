package log

import (
	"github.com/integration-system/isp-kit/log/file"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	IsInDevMode  bool
	FileOutput   *file.Output
	Sampling     *SamplingConfig
	Hooks        []func(entry zapcore.Entry) error
	InitialLevel Level
}

type SamplingConfig = zap.SamplingConfig

func DefaultConfig() *Config {
	return &Config{
		InitialLevel: InfoLevel,
	}
}

package log

import (
	"github.com/integration-system/isp-kit/log/file"
	"go.uber.org/zap"
)

type Config struct {
	IsInDevMode  bool
	FileOutput   *file.Output
	Sampling     *SamplingConfig
	InitialLevel Level
}

type SamplingConfig = zap.SamplingConfig

func DefaultConfig() *Config {
	return &Config{
		InitialLevel: InfoLevel,
	}
}

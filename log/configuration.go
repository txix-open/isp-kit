package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	IsInDevMode  bool
	OutputPaths  []string
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

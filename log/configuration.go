// Package log provides a logging adapter built on top of Uber's Zap library.
// It supports structured logging with configurable output, log levels, and hooks.
package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the configuration for the logging adapter.
type Config struct {
	// IsInDevMode enables development mode with human-friendly output.
	IsInDevMode bool
	// OutputPaths specifies the destinations for log output.
	OutputPaths []string
	// DisableDefaultOutput prevents logging to the default output.
	DisableDefaultOutput bool
	// SamplingConfig configures log message sampling.
	Sampling *SamplingConfig
	// Hooks is a slice of functions called for each log entry.
	Hooks []func(entry zapcore.Entry) error
	// InitialLevel sets the minimum log level.
	InitialLevel Level
}

// SamplingConfig is an alias for zap.SamplingConfig.
type SamplingConfig = zap.SamplingConfig

// DefaultConfig returns a Config with default settings.
func DefaultConfig() *Config {
	return &Config{
		InitialLevel: InfoLevel,
	}
}

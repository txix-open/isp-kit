package log

import (
	"github.com/txix-open/isp-kit/log/file"
)

// Option is a function that configures a Config.
type Option func(cfg *Config)

// WithDevelopmentMode enables development mode for human-friendly output.
func WithDevelopmentMode() Option {
	return func(a *Config) {
		a.IsInDevMode = true
	}
}

// WithFileOutput adds a file output with rotation configuration.
func WithFileOutput(fileOutput file.Output) Option {
	return func(a *Config) {
		a.OutputPaths = append(a.OutputPaths, file.ConfigToUrl(fileOutput).String())
	}
}

// WithDisableDefaultOutput disables the default output.
func WithDisableDefaultOutput() Option {
	return func(a *Config) {
		a.DisableDefaultOutput = true
	}
}

// WithLevel sets the initial log level.
func WithLevel(level Level) Option {
	return func(a *Config) {
		a.InitialLevel = level
	}
}

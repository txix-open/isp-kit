package log

import (
	"github.com/txix-open/isp-kit/log/file"
)

type Option func(cfg *Config)

func WithDevelopmentMode() Option {
	return func(a *Config) {
		a.IsInDevMode = true
	}
}

func WithFileOutput(fileOutput file.Output) Option {
	return func(a *Config) {
		a.OutputPaths = append(a.OutputPaths, file.ConfigToUrl(fileOutput).String())
	}
}

func WithDisableDefaultOutput() Option {
	return func(a *Config) {
		a.DisableDefaultOutput = true
	}
}

func WithLevel(level Level) Option {
	return func(a *Config) {
		a.InitialLevel = level
	}
}

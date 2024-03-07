package log

import (
	"github.com/integration-system/isp-kit/log/file"
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

func WithLevel(level Level) Option {
	return func(a *Config) {
		a.InitialLevel = level
	}
}

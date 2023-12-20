package app

import (
	"github.com/integration-system/isp-kit/config"
	"github.com/integration-system/isp-kit/log"
)

type LoggerConfigSupplier func(cfg *config.Config) log.Config

type Option func(c *Config)

type Config struct {
	LoggerConfigSupplier LoggerConfigSupplier
	ConfigOptions        []config.Option
}

func DefaultConfig() *Config {
	return &Config{
		LoggerConfigSupplier: func(cfg *config.Config) log.Config {
			return *log.DefaultConfig()
		},
	}
}

func WithConfigOptions(opts ...config.Option) Option {
	return func(c *Config) {
		c.ConfigOptions = append(c.ConfigOptions, opts...)
	}
}

func WithLoggerConfigSupplier(supplier LoggerConfigSupplier) Option {
	return func(c *Config) {
		c.LoggerConfigSupplier = supplier
	}
}

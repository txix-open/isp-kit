package app

import (
	"github.com/txix-open/isp-kit/config"
	"github.com/txix-open/isp-kit/log"
)

// LoggerConfigSupplier is a function that generates a logger configuration
// based on the application configuration.
type LoggerConfigSupplier func(cfg *config.Config) log.Config

// Option is a function that configures an Application configuration.
// Options follow the functional options pattern for flexible configuration.
type Option func(c *Config)

// Config holds the configuration for creating an Application.
type Config struct {
	// LoggerConfigSupplier generates the logger configuration from app config.
	LoggerConfigSupplier LoggerConfigSupplier
	// ConfigOptions are passed to the underlying configuration system.
	ConfigOptions []config.Option
}

// DefaultConfig returns a new Config with sensible defaults.
// The default logger configuration supplier returns a standard log.Config.
func DefaultConfig() *Config {
	return &Config{
		LoggerConfigSupplier: func(cfg *config.Config) log.Config {
			return *log.DefaultConfig()
		},
	}
}

// WithConfigOptions creates an Option that appends configuration options
// to the Application's config.
//
// Example:
//
//	app, err := app.New(
//		app.WithConfigOptions(
//			config.WithExtraSource(config.NewYamlConfig("config.yaml")),
//			config.WithEnvPrefix("APP"),
//		),
//	)
func WithConfigOptions(opts ...config.Option) Option {
	return func(c *Config) {
		c.ConfigOptions = append(c.ConfigOptions, opts...)
	}
}

// WithLoggerConfigSupplier creates an Option that sets the logger configuration
// supplier function.
//
// Example:
//
//	app, err := app.New(
//		app.WithLoggerConfigSupplier(func(cfg *config.Config) log.Config {
//			logCfg := *log.DefaultConfig()
//			logCfg.Level = log.DebugLevel
//			return logCfg
//		}),
//	)
func WithLoggerConfigSupplier(supplier LoggerConfigSupplier) Option {
	return func(c *Config) {
		c.LoggerConfigSupplier = supplier
	}
}

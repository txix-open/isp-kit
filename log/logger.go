package log

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Adapter is the main logging implementation built on Zap.
type Adapter struct {
	cfg    Config
	logger *zap.Logger
	level  zap.AtomicLevel
}

// New creates a new Adapter with the provided options.
func New(opts ...Option) (*Adapter, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return NewFromConfig(*config)
}

// NewFromConfig creates a new Adapter from a Config.
func NewFromConfig(config Config) (*Adapter, error) {
	cfg := zap.NewProductionConfig()
	if config.IsInDevMode {
		cfg = zap.NewDevelopmentConfig()
	}
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Sampling = config.Sampling

	if config.DisableDefaultOutput {
		cfg.OutputPaths = nil
	}

	if len(config.OutputPaths) > 0 {
		cfg.OutputPaths = append(cfg.OutputPaths, config.OutputPaths...)
	}

	level := zap.NewAtomicLevelAt(config.InitialLevel)
	cfg.Level = level

	var opts []zap.Option
	if len(config.Hooks) > 0 {
		opts = append(opts, zap.Hooks(config.Hooks...))
	}

	logger, err := cfg.Build(opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "build logger")
	}

	return &Adapter{
		cfg:    config,
		logger: logger,
		level:  level,
	}, nil
}

// Fatal logs a fatal-level message.
func (a *Adapter) Fatal(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, FatalLevel, message, fields...)
}

// Error logs an error-level message.
func (a *Adapter) Error(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, ErrorLevel, message, fields...)
}

// Warn logs a warning-level message.
func (a *Adapter) Warn(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, WarnLevel, message, fields...)
}

// Info logs an informational message.
func (a *Adapter) Info(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, InfoLevel, message, fields...)
}

// Debug logs a debug-level message.
func (a *Adapter) Debug(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, DebugLevel, message, fields...)
}

// Log writes a log entry at the specified level.
func (a *Adapter) Log(ctx context.Context, level Level, message any, fields ...Field) {
	entry := a.logger.Check(level, castString(message))
	if entry != nil {
		fields = append(fields, ContextLogValues(ctx)...)
		entry.Write(fields...)
	}
}

// SetLevel changes the current log level.
func (a *Adapter) SetLevel(level Level) {
	a.level.SetLevel(level)
}

// Enabled checks if the specified level is enabled.
func (a *Adapter) Enabled(level Level) bool {
	return a.level.Enabled(level)
}

// Sync flushes any buffered log entries.
func (a *Adapter) Sync() error {
	return a.logger.Sync()
}

// Config returns the current configuration.
func (a *Adapter) Config() Config {
	return a.cfg
}

// StdLoggerWithLevel wraps a Logger to provide a standard log.Logger interface.
func StdLoggerWithLevel(adapter Logger, level Level, withFields ...Field) *log.Logger {
	kitAdapter, ok := adapter.(*Adapter)
	if !ok {
		panic(fmt.Errorf("adapter must be a [%T], got [%T]", &Adapter{}, adapter)) // nolint:err113
	}
	logger := kitAdapter.logger.With(withFields...)
	stdLogger, err := zap.NewStdLogAt(logger, level)
	if err != nil {
		panic(err)
	}
	return stdLogger
}

// castString converts a value to a string representation.
func castString(v any) string {
	switch typed := v.(type) {
	case string:
		return typed
	case error:
		return typed.Error()
	default:
		return fmt.Sprintf("%v", typed)
	}
}

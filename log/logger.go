package log

import (
	"context"
	"fmt"
	"log"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Adapter struct {
	cfg    Config
	logger *zap.Logger
	level  zap.AtomicLevel
}

func New(opts ...Option) (*Adapter, error) {
	config := DefaultConfig()
	for _, opt := range opts {
		opt(config)
	}
	return NewFromConfig(*config)
}

func NewFromConfig(config Config) (*Adapter, error) {
	cfg := zap.NewProductionConfig()
	if config.IsInDevMode {
		cfg = zap.NewDevelopmentConfig()
	}
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Sampling = config.Sampling
	cfg.OutputPaths = append(cfg.OutputPaths, config.OutputPaths...)
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

func (a *Adapter) Fatal(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, FatalLevel, message, fields...)
}

func (a *Adapter) Error(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, ErrorLevel, message, fields...)
}

func (a *Adapter) Warn(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, WarnLevel, message, fields...)
}

func (a *Adapter) Info(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, InfoLevel, message, fields...)
}

func (a *Adapter) Debug(ctx context.Context, message any, fields ...Field) {
	a.Log(ctx, DebugLevel, message, fields...)
}

func (a *Adapter) Log(ctx context.Context, level Level, message any, fields ...Field) {
	entry := a.logger.Check(level, castString(message))
	if entry != nil {
		arr := append(ContextLogValues(ctx), fields...)
		entry.Write(arr...)
	}
}

func (a *Adapter) SetLevel(level Level) {
	a.level.SetLevel(level)
}

func (a *Adapter) Enabled(level Level) bool {
	return a.level.Enabled(level)
}

func (a *Adapter) Sync() error {
	return a.logger.Sync()
}

func (a *Adapter) Config() Config {
	return a.cfg
}

func StdLoggerWithLevel(adapter Logger, level Level, withFields ...Field) *log.Logger {
	kitAdapter, ok := adapter.(*Adapter)
	if !ok {
		panic(fmt.Errorf("adapter must be a [%T], got [%T]", &Adapter{}, adapter))
	}
	logger := kitAdapter.logger.With(withFields...)
	stdLogger, err := zap.NewStdLogAt(logger, level)
	if err != nil {
		panic(err)
	}
	return stdLogger
}

func castString(v any) string {
	switch typed := v.(type) {
	case string:
		return typed
	case error:
		return typed.Error()
	default:
		return fmt.Sprintf("%v", v)
	}
}

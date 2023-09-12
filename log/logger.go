package log

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Adapter struct {
	devMode      bool
	rotation     *Rotation
	initialLevel Level

	logger *zap.Logger
	level  zap.AtomicLevel
}

func New(opts ...Option) (*Adapter, error) {
	a := &Adapter{
		devMode:      false,
		rotation:     nil,
		initialLevel: InfoLevel,
	}
	for _, opt := range opts {
		opt(a)
	}

	cfg := zap.NewProductionConfig()
	if a.devMode {
		cfg = zap.NewDevelopmentConfig()
	}
	cfg.DisableStacktrace = true
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if a.rotation != nil {
		outputUrl := rotationToUrl(*a.rotation)
		cfg.OutputPaths = append(cfg.OutputPaths, outputUrl.String())
	}
	level := zap.NewAtomicLevelAt(a.initialLevel)
	cfg.Level = level
	logger, err := cfg.Build()
	if err != nil {
		return nil, errors.WithMessage(err, "build logger")
	}
	a.logger = logger
	a.level = level

	return a, nil
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
	entry := a.logger.Check(level, fmt.Sprintf("%v", message))
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

func (a *Adapter) Close() error {
	return a.logger.Sync()
}

package sentry

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"go.uber.org/zap/zapcore"
)

var (
	logLevelMapping = map[log.Level]sentry.Level{
		log.FatalLevel: sentry.LevelFatal,
		log.ErrorLevel: sentry.LevelError,
		log.WarnLevel:  sentry.LevelWarning,
		log.InfoLevel:  sentry.LevelInfo,
		log.DebugLevel: sentry.LevelDebug,
	}
)

type Logger struct {
	delegate        log.Logger
	hub             Hub
	supportedLevels map[log.Level]sentry.Level
}

func WrapErrorLogger(logger log.Logger, hub Hub) Logger {
	return WrapLogger(
		logger,
		hub,
		[]log.Level{log.ErrorLevel},
	)
}

func WrapLogger(logger log.Logger, hub Hub, supportedLevels []log.Level) Logger {
	levels := make(map[log.Level]sentry.Level)
	for _, level := range supportedLevels {
		levels[level] = logLevelMapping[level]
	}

	return Logger{
		delegate:        logger,
		hub:             hub,
		supportedLevels: levels,
	}
}

func (s Logger) Error(ctx context.Context, message any, fields ...log.Field) {
	s.delegate.Error(ctx, message, fields...)
	s.log(log.ErrorLevel, ctx, message, fields...)
}

func (s Logger) Warn(ctx context.Context, message any, fields ...log.Field) {
	s.delegate.Warn(ctx, message, fields...)
	s.log(log.WarnLevel, ctx, message, fields...)
}

func (s Logger) Info(ctx context.Context, message any, fields ...log.Field) {
	s.delegate.Info(ctx, message, fields...)
	s.log(log.InfoLevel, ctx, message, fields...)
}

func (s Logger) Debug(ctx context.Context, message any, fields ...log.Field) {
	s.delegate.Debug(ctx, message, fields...)
	s.log(log.DebugLevel, ctx, message, fields...)
}

func (s Logger) log(level log.Level, ctx context.Context, message any, fields ...log.Field) {
	sentryLevel, isSupported := s.supportedLevels[level]
	if !isSupported {
		return
	}

	_, isNoopHub := s.hub.(NoopHub)
	if isNoopHub {
		return
	}

	event := EventFromLog(sentryLevel, ctx, message, fields...)
	s.hub.CatchEvent(ctx, event)
}

func EventFromLog(
	level sentry.Level,
	ctx context.Context,
	message any,
	fields ...log.Field,
) *sentry.Event {
	errInLog, _ := message.(error)

	extra := make(map[string]any, len(fields))
	for _, field := range log.ContextLogValues(ctx) {
		value := fieldToAny(field)
		extra[field.Key] = value
	}
	for _, field := range fields {
		value := fieldToAny(field)
		extra[field.Key] = value
		errField := errorFromField(field)
		if errInLog == nil {
			errInLog = errField
		}
	}

	requestId := requestid.FromContext(ctx)
	if requestId != "" {
		extra["requestId"] = requestId
	}

	event := &sentry.Event{
		Extra:     extra,
		Level:     level,
		Message:   fmt.Sprintf("%v", message),
		Timestamp: time.Now(),
	}
	if errInLog != nil {
		SetException(event, errInLog)
	}

	enrichment, ok := ctx.Value(contextKeyValue).(EventEnrichment)
	if ok {
		enrichment(event)
	}

	return event
}

// SetException
// Compare to sentry.Event.SetException, we don't add local stacktrace if there is no stack in error
func SetException(e *sentry.Event, exception error) {
	err := exception
	if err == nil {
		return
	}

	for i := 0; i < 10 && err != nil; i++ {
		e.Exception = append(e.Exception, sentry.Exception{
			Value:      err.Error(),
			Type:       reflect.TypeOf(err).String(),
			Stacktrace: sentry.ExtractStacktrace(err),
		})
		switch previous := err.(type) {
		case interface{ Unwrap() error }:
			err = previous.Unwrap()
		case interface{ Cause() error }:
			err = previous.Cause()
		default:
			err = nil
		}
	}

	slices.Reverse(e.Exception)
}

func fieldToAny(field log.Field) any {
	switch {
	case field.String != "":
		return field.String
	case field.Type == zapcore.BoolType:
		return field.Integer == 1
	case field.Type == zapcore.ByteStringType:
		return string(field.Interface.([]byte))
	case field.Type == zapcore.StringerType:
		return field.Interface.(fmt.Stringer).String()
	case field.Type == zapcore.DurationType:
		return time.Duration(field.Integer).String()
	case field.Type == zapcore.TimeType:
		return time.Unix(0, field.Integer).In(field.Interface.(*time.Location)).Format("2006-01-02T15:04:05.000Z0700")
	case field.Interface != nil:
		return fmt.Sprintf("%v", field.Interface)
	default:
		return field.Integer
	}
}

func errorFromField(field log.Field) error {
	if field.Type == zapcore.ErrorType {
		return field.Interface.(error)
	}
	return nil
}

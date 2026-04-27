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

// logLevelMapping maps log levels to Sentry event levels.
var (
	logLevelMapping = map[log.Level]sentry.Level{
		log.FatalLevel: sentry.LevelFatal,
		log.ErrorLevel: sentry.LevelError,
		log.WarnLevel:  sentry.LevelWarning,
		log.InfoLevel:  sentry.LevelInfo,
		log.DebugLevel: sentry.LevelDebug,
	}
)

// Logger wraps a log.Logger and forwards log events to Sentry.
// It provides structured logging with automatic error capture and context enrichment.
type Logger struct {
	delegate        log.Logger
	hub             Hub
	supportedLevels map[log.Level]sentry.Level
}

// WrapErrorLogger creates a Logger that captures only error-level events.
// It delegates all logging to the underlying logger and sends errors to Sentry.
func WrapErrorLogger(logger log.Logger, hub Hub) Logger {
	return WrapLogger(
		logger,
		hub,
		[]log.Level{log.ErrorLevel},
	)
}

// WrapLogger creates a Logger that captures events at the specified log levels.
// Only the levels in supportedLevels will be forwarded to Sentry.
// The returned Logger is safe for concurrent use.
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

// Error logs an error message and forwards it to Sentry if ErrorLevel is supported.
// It delegates to the underlying logger and captures the error with full context.
func (s Logger) Error(ctx context.Context, message any, fields ...log.Field) {
	s.delegate.Error(ctx, message, fields...)
	s.log(log.ErrorLevel, ctx, message, fields...)
}

// Warn logs a warning message and forwards it to Sentry if WarnLevel is supported.
func (s Logger) Warn(ctx context.Context, message any, fields ...log.Field) {
	s.delegate.Warn(ctx, message, fields...)
	s.log(log.WarnLevel, ctx, message, fields...)
}

// Info logs an informational message and forwards it to Sentry if InfoLevel is supported.
func (s Logger) Info(ctx context.Context, message any, fields ...log.Field) {
	s.delegate.Info(ctx, message, fields...)
	s.log(log.InfoLevel, ctx, message, fields...)
}

// Debug logs a debug message and forwards it to Sentry if DebugLevel is supported.
func (s Logger) Debug(ctx context.Context, message any, fields ...log.Field) {
	s.delegate.Debug(ctx, message, fields...)
	s.log(log.DebugLevel, ctx, message, fields...)
}

// log forwards a log event to Sentry if the level is supported and the hub is not a NoopHub.
// It extracts error information from fields and enriches the event with context.
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

// EventFromLog creates a Sentry event from a log entry.
// It extracts the error from the message or fields, enriches the event with context values,
// and applies any event enrichment functions from the context.
// The returned event includes the request ID if present in the context.
func EventFromLog(
	level sentry.Level,
	ctx context.Context,
	message any,
	fields ...log.Field,
) *sentry.Event {
	errInLog, _ := message.(error)

	tags := make(map[string]string, len(fields))
	for _, field := range log.ContextLogValues(ctx) {
		value := fieldToString(field)
		tags[field.Key] = value
	}
	for _, field := range fields {
		value := fieldToString(field)
		tags[field.Key] = value
		errField := errorFromField(field)
		if errInLog == nil {
			errInLog = errField
		}
	}

	requestId := requestid.FromContext(ctx)
	if requestId != "" {
		tags["requestId"] = requestId
	}

	event := &sentry.Event{
		Tags:      tags,
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

// SetException extracts the error chain from an error and sets it on the Sentry event.
// Unlike the Sentry SDK's SetException, it only includes stack traces where they exist.
// The exception list is reversed to show the root cause first.
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
		// nolint:errorlint
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

// nolint:forcetypeassert
// fieldToString converts a log.Field to a Go string.
// It handles various field types including strings, booleans, times, durations, and stringers.
func fieldToString(field log.Field) string {
	switch {
	case field.String != "":
		return field.String
	case field.Type == zapcore.BoolType:
		return fmt.Sprintf("%v", field.Integer == 1)
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
		return fmt.Sprintf("%v", field.Integer)
	}
}

// errorFromField extracts an error from a log.Field if it is an error type.
// Returns nil if the field is not an error.
func errorFromField(field log.Field) error {
	if field.Type == zapcore.ErrorType {
		return field.Interface.(error) // nolint:forcetypeassert
	}
	return nil
}

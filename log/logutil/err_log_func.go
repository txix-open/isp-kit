// Package logutil provides utility functions for logging.
package logutil

import (
	"context"

	"github.com/txix-open/isp-kit/log"
)

// LogLevelSpecifier is an interface for errors that specify their log level.
type LogLevelSpecifier interface {
	LogLevel() log.Level
}

// LogLevelForError determines the log level for an error.
// It returns ErrorLevel by default, or a custom level if the error implements LogLevelSpecifier.
func LogLevelForError(err error) log.Level {
	logLevel := log.ErrorLevel
	specifier, ok := err.(LogLevelSpecifier)
	if ok {
		logLevel = specifier.LogLevel()
	}
	return logLevel
}

// LogLevelFuncForError returns the appropriate logging function for an error.
// It uses LogLevelForError to determine the log level and returns the corresponding Logger method.
func LogLevelFuncForError(err error, logger log.Logger) func(ctx context.Context, message any, fields ...log.Field) {
	logLevel := LogLevelForError(err)
	switch logLevel {
	case log.ErrorLevel:
		return logger.Error
	case log.WarnLevel:
		return logger.Warn
	case log.InfoLevel:
		return logger.Info
	case log.DebugLevel:
		return logger.Debug
	default:
		return logger.Error
	}
}

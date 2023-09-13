package logutil

import (
	"context"

	"github.com/integration-system/isp-kit/log"
)

type LogLevelSpecifier interface {
	LogLevel() log.Level
}

func LogLevelFuncForError(err error, logger log.Logger) func(ctx context.Context, message any, fields ...log.Field) {
	logLevel := log.ErrorLevel
	specifier, ok := err.(LogLevelSpecifier)
	if ok {
		logLevel = specifier.LogLevel()
	}

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

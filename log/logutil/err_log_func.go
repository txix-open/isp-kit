package logutil

import (
	"context"

	"gitlab.txix.ru/isp/isp-kit/log"
)

type LogLevelSpecifier interface {
	LogLevel() log.Level
}

func LogLevelForError(err error) log.Level {
	logLevel := log.ErrorLevel
	specifier, ok := err.(LogLevelSpecifier)
	if ok {
		logLevel = specifier.LogLevel()
	}
	return logLevel
}

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

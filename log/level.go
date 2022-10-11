package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	FatalLevel = zap.FatalLevel
	ErrorLevel = zap.ErrorLevel
	InfoLevel  = zap.InfoLevel
	DebugLevel = zap.DebugLevel
)

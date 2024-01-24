package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level = zapcore.Level

const (
	FatalLevel = zap.FatalLevel
	ErrorLevel = zap.ErrorLevel
	WarnLevel  = zap.WarnLevel
	InfoLevel  = zap.InfoLevel
	DebugLevel = zap.DebugLevel
)

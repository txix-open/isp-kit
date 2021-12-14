package log

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level zapcore.Level

func (l *Level) UnmarshalText(text []byte) error {
	if l == nil {
		return errors.New("can't unmarshal a nil *Level")
	}
	return (*zapcore.Level)(l).UnmarshalText(text)
}

func (l Level) MarshalText() (text []byte, err error) {
	return zapcore.Level(l).MarshalText()
}

const (
	FatalLevel = Level(zap.FatalLevel)
	ErrorLevel = Level(zap.ErrorLevel)
	InfoLevel  = Level(zap.InfoLevel)
	DebugLevel = Level(zap.DebugLevel)
)

package kafkax

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/txix-open/isp-kit/log"
)

type Logger struct {
	ctx    context.Context //nolint:containedctx
	name   string
	level  kgo.LogLevel
	logger log.Logger
}

func NewLogger(ctx context.Context, name string, level kgo.LogLevel, logger log.Logger) *Logger {
	return &Logger{
		ctx:    ctx,
		name:   name,
		level:  level,
		logger: logger,
	}
}

func (l *Logger) Level() kgo.LogLevel {
	return l.level
}

func (l *Logger) Log(level kgo.LogLevel, msg string, keyvals ...any) {
	logMsg := fmt.Sprintf("%s: %s", l.name, fmt.Sprintf(msg, keyvals...))
	// nolint:exhaustive
	switch level {
	case kgo.LogLevelError:
		l.logger.Error(l.ctx, logMsg)
	case kgo.LogLevelWarn:
		l.logger.Warn(l.ctx, logMsg)
	case kgo.LogLevelInfo:
		l.logger.Info(l.ctx, logMsg)
	case kgo.LogLevelDebug:
		l.logger.Debug(l.ctx, logMsg)
	}
}

package kafkax

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/txix-open/isp-kit/log"
)

// Logger is an adapter that bridges franz-go's logging interface to the
// application's logger. It prefixes log messages with a configurable name
// and maps franz-go log levels to the application's log levels.
type Logger struct {
	ctx    context.Context //nolint:containedctx
	name   string
	level  kgo.LogLevel
	logger log.Logger
}

// NewLogger creates a new Logger instance with the provided configuration.
func NewLogger(ctx context.Context, name string, level kgo.LogLevel, logger log.Logger) *Logger {
	return &Logger{
		ctx:    ctx,
		name:   name,
		level:  level,
		logger: logger,
	}
}

// Level returns the logger's configured log level.
func (l *Logger) Level() kgo.LogLevel {
	return l.level
}

// Log writes a log message at the specified level. The message is prefixed
// with the logger's name and forwarded to the underlying logger.
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

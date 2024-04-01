package migration

import (
	"log"

	kitLog "github.com/txix-open/isp-kit/log"
)

type logger struct {
	stdLogger *log.Logger
}

func (l logger) Fatalf(format string, v ...interface{}) {
	l.stdLogger.Fatalf(format, v...)
}

func (l logger) Printf(format string, v ...interface{}) {
	l.stdLogger.Printf(format, v...)
}

func newLogger(kitLogger kitLog.Logger) logger {
	return logger{
		stdLogger: kitLog.StdLoggerWithLevel(
			kitLogger,
			kitLog.InfoLevel,
			kitLog.String("worker", "goose_db_migration"),
		),
	}
}

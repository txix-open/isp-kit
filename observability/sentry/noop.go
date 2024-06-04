package sentry

import (
	"context"

	"github.com/getsentry/sentry-go"
	"gitlab.txix.ru/isp/isp-kit/log"
)

type NoopHub struct {
}

func NewNoopHub() NoopHub {
	return NoopHub{}
}

func (n NoopHub) CatchError(ctx context.Context, err error, level log.Level) {
}

func (n NoopHub) CatchEvent(ctx context.Context, event *sentry.Event) {

}

func (n NoopHub) Flush() {

}

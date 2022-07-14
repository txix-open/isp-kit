package healthcheck

import (
	"context"
)

type Checker interface {
	Healthcheck(ctx context.Context) error
}

type CheckerFunc func(ctx context.Context) error

func (r CheckerFunc) Healthcheck(ctx context.Context) error {
	return r(ctx)
}

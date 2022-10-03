package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v4"
)

type ExponentialBackoff struct {
	maxElapsedTime time.Duration
}

func NewExponentialBackoff(maxElapsedTime time.Duration) ExponentialBackoff {
	return ExponentialBackoff{
		maxElapsedTime: maxElapsedTime,
	}
}

func (e ExponentialBackoff) Do(ctx context.Context, operation func() error) error {
	exp := backoff.NewExponentialBackOff()
	exp.MaxElapsedTime = e.maxElapsedTime
	withCtx := backoff.WithContext(exp, ctx)
	return backoff.Retry(operation, withCtx)
}

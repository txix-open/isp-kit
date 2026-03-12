package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v5"
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
	backOff := backoff.WithBackOff(backoff.NewExponentialBackOff())
	maxElapsedTime := backoff.WithMaxElapsedTime(e.maxElapsedTime)
	_, err := backoff.Retry[any](ctx, func() (any, error) {
		return nil, operation()
	}, backOff, maxElapsedTime)
	return err
}

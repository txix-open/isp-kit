// Package retry provides utilities for performing operations with retry logic.
// It wraps the cenkalti/backoff library to support exponential backoff strategies
// with context cancellation.
package retry

import (
	"context"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// ExponentialBackoff configures and executes operations using an exponential backoff strategy.
type ExponentialBackoff struct {
	maxElapsedTime time.Duration
}

// NewExponentialBackoff creates a new ExponentialBackoff instance with the specified
// maximum elapsed time limit.
func NewExponentialBackoff(maxElapsedTime time.Duration) ExponentialBackoff {
	return ExponentialBackoff{
		maxElapsedTime: maxElapsedTime,
	}
}

// Do executes the provided operation, retrying it on failure using exponential backoff.
// The function returns when the operation succeeds, the context is cancelled, or the
// maxElapsedTime limit is reached. Returns the last error if all attempts fail.
func (e ExponentialBackoff) Do(ctx context.Context, operation func() error) error {
	backOff := backoff.WithBackOff(backoff.NewExponentialBackOff())
	maxElapsedTime := backoff.WithMaxElapsedTime(e.maxElapsedTime)
	_, err := backoff.Retry[any](ctx, func() (any, error) {
		return nil, operation()
	}, backOff, maxElapsedTime)
	return err
}

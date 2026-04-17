package httpcli

import (
	"context"
	"fmt"
	"net/http"
)

// Retryer defines the interface for retry logic.
// Implementations control when and how many times to retry failed operations.
type Retryer interface {
	Do(ctx context.Context, f func() error) error
}

// RetryCondition is a function that determines whether to retry based on
// the error and response. Returns nil if no retry is needed.
type RetryCondition func(err error, response *Response) error

// retryOptions holds retry configuration for a request.
type retryOptions struct {
	condition RetryCondition
	retrier   Retryer
}

// nolint:gochecknoglobals
var (
	noRetries = &retryOptions{
		condition: func(err error, response *Response) error {
			return err
		},
		retrier: noRetryer{},
	}
)

// noRetryer is a retryer that never retries.
type noRetryer struct {
}

// Do executes the function exactly once without retry logic.
func (n noRetryer) Do(_ context.Context, f func() error) error {
	return f()
}

// NoRetries returns a retry condition and retryer that disable retry behavior.
//
// nolint:ireturn
func NoRetries() (RetryCondition, Retryer) {
	return noRetries.condition, noRetries.retrier
}

// IfErrorOr5XXStatus returns a retry condition that retries on network errors
// or 5xx status codes.
func IfErrorOr5XXStatus() RetryCondition {
	return func(err error, response *Response) error {
		if err != nil {
			return err
		}
		if response.Raw.StatusCode >= http.StatusInternalServerError && response.Raw.StatusCode <= 599 {
			return fmt.Errorf("status code %d", response.Raw.StatusCode) // nolint:err113
		}
		return nil
	}
}

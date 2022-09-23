package httpcli

import (
	"context"
	"fmt"
	"net/http"
)

type Retryer interface {
	Do(ctx context.Context, f func() error) error
}

type RetryCondition func(err error, response *Response) error

type retryOptions struct {
	condition RetryCondition
	retrier   Retryer
}

var (
	noRetries = &retryOptions{
		condition: func(err error, response *Response) error {
			return err
		},
		retrier: noRetryer{},
	}
)

type noRetryer struct {
}

func (n noRetryer) Do(_ context.Context, f func() error) error {
	return f()
}

func NoRetries() (RetryCondition, Retryer) {
	return noRetries.condition, noRetries.retrier
}

func IfErrorOr5XXStatus() RetryCondition {
	return func(err error, response *Response) error {
		if err != nil {
			return err
		}
		if response.Raw.StatusCode >= http.StatusInternalServerError && response.Raw.StatusCode <= 599 {
			return fmt.Errorf("status code %d", response.Raw.StatusCode)
		}
		return nil
	}
}

package handler

import (
	"time"
)

type Result struct {
	Commit     bool
	Retry      bool
	RetryError error
	RetryAfter time.Duration
}

func Commit() Result {
	return Result{Commit: true}
}

func Retry(after time.Duration, err error) Result {
	return Result{Retry: true, RetryAfter: after, RetryError: err}
}

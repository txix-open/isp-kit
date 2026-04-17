package handler

import (
	"time"
)

// Result defines the outcome of message processing, indicating whether to
// commit the offset, retry processing, or skip the message.
type Result struct {
	Commit     bool
	Retry      bool
	RetryError error
	RetryAfter time.Duration
}

// Commit returns a Result indicating the message should be committed.
func Commit() Result {
	return Result{Commit: true}
}

// Nothing returns a Result indicating the message should be skipped without
// committing.
func Nothing() Result {
	return Result{}
}

// Retry returns a Result indicating the message should be retried after the
// specified duration. The error will be logged for diagnostic purposes.
func Retry(after time.Duration, err error) Result {
	return Result{Retry: true, RetryAfter: after, RetryError: err}
}

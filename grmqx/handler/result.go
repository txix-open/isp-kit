package handler

import (
	"time"
)

// Result represents the outcome of message processing.
type Result struct {
	// Ack indicates the message should be acknowledged (successfully processed).
	Ack bool
	// Requeue indicates the message should be requeued after a delay.
	Requeue bool
	// RequeueTimeout specifies the delay before requeuing the message.
	RequeueTimeout time.Duration
	// Retry indicates the message should be retried using the retry policy.
	Retry bool
	// MoveToDlq indicates the message should be moved to the dead letter queue.
	MoveToDlq bool
	// Err contains the error that occurred during processing.
	Err error
}

// Ack creates a Result indicating successful message processing.
func Ack() Result {
	return Result{
		Ack: true,
	}
}

// Requeue creates a Result indicating the message should be requeued after a delay.
//
// Deprecated: use Retry and RetryPolicy instead.
func Requeue(after time.Duration, err error) Result {
	return Result{
		Requeue:        true,
		RequeueTimeout: after,
		Err:            err,
	}
}

// MoveToDlq creates a Result indicating the message should be moved to the dead letter queue.
// If no DLQ is configured, the message will be dropped.
func MoveToDlq(err error) Result {
	return Result{
		MoveToDlq: true,
		Err:       err,
	}
}

// Retry creates a Result indicating the message should be retried.
func Retry(err error) Result {
	return Result{
		Retry: true,
		Err:   err,
	}
}

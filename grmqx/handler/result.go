package handler

import (
	"time"
)

type Result struct {
	Ack            bool
	Requeue        bool
	RequeueTimeout time.Duration
	Retry          bool
	MoveToDlq      bool
	Err            error
}

func Ack() Result {
	return Result{
		Ack: true,
	}
}

// Requeue
// Deprecated: use Retry and RetryPolicy instead
func Requeue(after time.Duration, err error) Result {
	return Result{
		Requeue:        true,
		RequeueTimeout: after,
		Err:            err,
	}
}

// MoveToDlq
// if there is no DLQ, the message will be dropped
func MoveToDlq(err error) Result {
	return Result{
		MoveToDlq: true,
		Err:       err,
	}
}

func Retry(err error) Result {
	return Result{
		Retry: true,
		Err:   err,
	}
}

package handler

import (
	"time"
)

type Result struct {
	Complete   bool
	Err        error
	MoveToDlq  bool
	Retry      bool
	RetryDelay time.Duration

	Reschedule      bool
	RescheduleDelay time.Duration
	OverrideArg     bool
	Arg             []byte
}

func Complete() Result {
	return Result{Complete: true}
}

func Retry(after time.Duration, err error) Result {
	return Result{Retry: true, RetryDelay: after, Err: err}
}

func MoveToDlq(err error) Result {
	return Result{MoveToDlq: true, Err: err}
}

func Reschedule(after time.Duration) Result {
	return Result{Reschedule: true, RescheduleDelay: after}
}

func RescheduleWithArg(after time.Duration, arg []byte) Result {
	return Result{Reschedule: true, OverrideArg: true, RescheduleDelay: after, Arg: arg}
}

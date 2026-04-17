// Package panic_recovery provides utilities for recovering from panics in Go applications.
// It captures the panic value, formats it with a stack trace, and passes the result
// to a provided callback function for handling.
//
// The primary function, Recover, is intended to be used in a defer statement to
// catch panics that occur in the calling goroutine.
package panic_recovery

import (
	"runtime"

	"github.com/pkg/errors"
)

const (
	panicStackLength = 4 << 10
)

// Recover captures a panic, formats it with a stack trace, and passes the
// resulting error to the provided callback function. It should typically be
// called via defer to handle panics in the current goroutine.
//
// The formatRes function is invoked with a wrapped error containing the panic
// value and the current goroutine's stack trace. If no panic has occurred,
// Recover returns immediately without calling formatRes.
//
// Example usage:
//
//	defer Recover(func(err error) {
//	    log.Printf("Recovered from panic: %v", err)
//	})
func Recover(formatRes func(err error)) {
	r := recover()
	if r == nil {
		return
	}

	var err error
	recovered, ok := r.(error)
	if ok {
		err = recovered
	} else {
		err = errors.Errorf("%v", recovered)
	}
	stack := make([]byte, panicStackLength)
	length := runtime.Stack(stack, false)

	formatRes(errors.Errorf("[PANIC RECOVER] %v %s\n", err, stack[:length]))
}

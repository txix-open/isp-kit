package panic_recovery

import (
	"runtime"

	"github.com/pkg/errors"
)

const (
	panicStackLength = 4 << 10
)

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

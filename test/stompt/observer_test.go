package stompt_test

import (
	"sync/atomic"

	"github.com/txix-open/isp-kit/stompx/consumer"
)

type ErrorCountingObserver struct {
	ErrorCount atomic.Int32
}

func (o *ErrorCountingObserver) Error(c *consumer.Consumer, err error) {
	o.ErrorCount.Add(1)
}

func (o *ErrorCountingObserver) CloseStart(c *consumer.Consumer) {
}

func (o *ErrorCountingObserver) CloseDone(c *consumer.Consumer) {
}

func (o *ErrorCountingObserver) BeginConsuming(c *consumer.Consumer) {
}

func (o *ErrorCountingObserver) ErrorCounter() {
}

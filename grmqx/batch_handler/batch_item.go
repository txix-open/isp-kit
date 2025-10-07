package batch_handler

import (
	"context"

	"github.com/txix-open/grmq/consumer"
)

// nolint:containedctx
type BatchItem struct {
	Context  context.Context
	Delivery *consumer.Delivery
	Result   Result
}

func (b *BatchItem) Ack() {
	b.Result = Result{
		Ack: true,
	}
}

func (b *BatchItem) MoveToDlq(err error) {
	b.Result = Result{
		MoveToDlq: true,
		Err:       err,
	}
}

func (b *BatchItem) Retry(err error) {
	b.Result = Result{
		Retry: true,
		Err:   err,
	}
}

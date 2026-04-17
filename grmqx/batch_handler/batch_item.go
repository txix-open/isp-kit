package batch_handler

import (
	"context"

	"github.com/txix-open/grmq/consumer"
)

// BatchItem represents a single message in a batch with its processing result.
type BatchItem struct {
	// Context contains the request context.
	Context context.Context
	// Delivery contains the raw message delivery from the consumer.
	Delivery *consumer.Delivery
	// Result stores the processing outcome.
	Result Result
}

// Ack sets the result to indicate successful message acknowledgment.
func (b *BatchItem) Ack() {
	b.Result = Result{
		Ack: true,
	}
}

// MoveToDlq sets the result to indicate the message should be moved to the DLQ.
func (b *BatchItem) MoveToDlq(err error) {
	b.Result = Result{
		MoveToDlq: true,
		Err:       err,
	}
}

// Retry sets the result to indicate the message should be retried.
func (b *BatchItem) Retry(err error) {
	b.Result = Result{
		Retry: true,
		Err:   err,
	}
}

// BatchItems is a slice of batch items.
type BatchItems []*BatchItem

// AckAll sets all items to be acknowledged.
func (bs BatchItems) AckAll() {
	for _, b := range bs {
		b.Result = Result{
			Ack: true,
		}
	}
}

// MoveToDlqAll sets all items to be moved to the DLQ.
func (bs BatchItems) MoveToDlqAll(err error) {
	for _, b := range bs {
		b.Result = Result{
			MoveToDlq: true,
			Err:       err,
		}
	}
}

// RetryAll sets all items to be retried.
func (bs BatchItems) RetryAll(err error) {
	for _, b := range bs {
		b.Result = Result{
			Retry: true,
			Err:   err,
		}
	}
}

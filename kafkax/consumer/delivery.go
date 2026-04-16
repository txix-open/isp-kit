package consumer

import (
	"context"
	"github.com/pkg/errors"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	// ErrDeliveryAlreadyHandled is returned when attempting to commit a delivery
	// that has already been handled.
	ErrDeliveryAlreadyHandled = errors.New("delivery already handled")
)

// Donner defines an interface for signaling when message processing is complete.
type Donner interface {
	Done()
}

// Delivery represents a consumed message from Kafka. It provides methods to
// access the message source, commit offsets, and signal completion.
type Delivery struct {
	donner          Donner
	client          *kgo.Client
	source          *kgo.Record
	handled         bool
	consumerGroupId string
}

// NewDelivery creates a new Delivery instance with the provided configuration.
func NewDelivery(donner Donner, client *kgo.Client, source *kgo.Record, consumerGroupId string) *Delivery {
	return &Delivery{
		donner:          donner,
		client:          client,
		source:          source,
		consumerGroupId: consumerGroupId,
	}
}

// Source returns the underlying Kafka record containing the consumed message.
func (d *Delivery) Source() *kgo.Record {
	return d.source
}

// ConsumerGroupId returns the consumer group ID associated with this delivery.
func (d *Delivery) ConsumerGroupId() string {
	return d.consumerGroupId
}

// Commit commits the message offset to Kafka and signals completion. Returns
// an error if the delivery has already been handled. Only one of Commit or
// Done should be called per delivery.
func (d *Delivery) Commit(ctx context.Context) error {
	if d.handled {
		return ErrDeliveryAlreadyHandled
	}

	defer d.donner.Done()
	d.handled = true

	err := d.client.CommitRecords(ctx, d.source)
	if err != nil {
		return errors.WithMessage(err, "commit messages")
	}

	return nil
}

// Done signals that message processing is complete without committing the
// offset. This is typically used when the message should be skipped or
// reprocessed later.
func (d *Delivery) Done() {
	d.donner.Done()
}

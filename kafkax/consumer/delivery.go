package consumer

import (
	"context"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
)

var (
	ErrDeliveryAlreadyHandled = errors.New("delivery already handled")
)

type Donner interface {
	Done()
}

type Delivery struct {
	donner          Donner
	reader          *kafka.Reader
	source          *kafka.Message
	handled         bool
	consumerGroupId string
}

func NewDelivery(donner Donner, reader *kafka.Reader, source *kafka.Message, consumerGroupId string) *Delivery {
	return &Delivery{
		donner:          donner,
		reader:          reader,
		source:          source,
		consumerGroupId: consumerGroupId,
	}
}

func (d *Delivery) Source() *kafka.Message {
	return d.source
}

func (d *Delivery) ConsumerGroupId() string {
	return d.consumerGroupId
}

func (d *Delivery) Commit(ctx context.Context) error {
	if d.handled {
		return ErrDeliveryAlreadyHandled
	}

	defer d.donner.Done()
	d.handled = true

	err := d.reader.CommitMessages(ctx, *d.source)
	if err != nil {
		return errors.WithMessage(err, "commit messages")
	}

	return nil
}

func (d *Delivery) Done() {
	d.donner.Done()
}

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
	donner  Donner
	reader  *kafka.Reader
	source  *kafka.Message
	handled bool
}

func NewDelivery(donner Donner, reader *kafka.Reader, source *kafka.Message) *Delivery {
	return &Delivery{
		donner: donner,
		reader: reader,
		source: source,
	}
}

func (d *Delivery) Source() *kafka.Message {
	return d.source
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

package consumer

import (
	"context"
	"github.com/pkg/errors"
	"github.com/twmb/franz-go/pkg/kgo"
)

var (
	ErrDeliveryAlreadyHandled = errors.New("delivery already handled")
)

type Donner interface {
	Done()
}

type Delivery struct {
	donner          Donner
	client          *kgo.Client
	source          *kgo.Record
	handled         bool
	consumerGroupId string
}

func NewDelivery(donner Donner, client *kgo.Client, source *kgo.Record, consumerGroupId string) *Delivery {
	return &Delivery{
		donner:          donner,
		client:          client,
		source:          source,
		consumerGroupId: consumerGroupId,
	}
}

func (d *Delivery) Source() *kgo.Record {
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

	err := d.client.CommitRecords(ctx, d.source)
	if err != nil {
		return errors.WithMessage(err, "commit messages")
	}

	return nil
}

func (d *Delivery) Done() {
	d.donner.Done()
}

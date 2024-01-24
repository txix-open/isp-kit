package consumer

import (
	"github.com/go-stomp/stomp/v3"
	"github.com/pkg/errors"
)

var (
	ErrDeliveryAlreadyHandled = errors.New("delivery already handled")
)

type Donner interface {
	Done()
}

type Delivery struct {
	donner  Donner
	conn    *stomp.Conn
	source  *stomp.Message
	handled bool
}

func NewDelivery(donner Donner, conn *stomp.Conn, source *stomp.Message) *Delivery {
	return &Delivery{
		donner: donner,
		conn:   conn,
		source: source,
	}
}

func (d *Delivery) Source() *stomp.Message {
	return d.source
}

func (d *Delivery) Ack() error {
	if d.handled {
		return ErrDeliveryAlreadyHandled
	}

	defer d.donner.Done()
	d.handled = true

	err := d.conn.Ack(d.source)
	if err != nil {
		return errors.WithMessage(err, "ack delivery")
	}
	return nil
}

func (d *Delivery) Nack() error {
	if d.handled {
		return ErrDeliveryAlreadyHandled
	}

	defer d.donner.Done()
	d.handled = true

	err := d.conn.Nack(d.source)
	if err != nil {
		return errors.WithMessage(err, "nack delivery")
	}
	return nil
}

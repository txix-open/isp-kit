package consumer

import (
	"github.com/go-stomp/stomp/v3"
	"github.com/pkg/errors"
)

// ErrDeliveryAlreadyHandled is returned when attempting to acknowledge a delivery that has already been processed.
var (
	ErrDeliveryAlreadyHandled = errors.New("delivery already handled")
)

// Donner defines an interface for signaling completion.
type Donner interface {
	Done()
}

// Delivery represents a message delivery that can be acknowledged or negatively acknowledged.
type Delivery struct {
	donner  Donner
	conn    *stomp.Conn
	source  *stomp.Message
	handled bool
}

// NewDelivery creates a new delivery instance.
func NewDelivery(donner Donner, conn *stomp.Conn, source *stomp.Message) *Delivery {
	return &Delivery{
		donner: donner,
		conn:   conn,
		source: source,
	}
}

// Source returns the underlying STOMP message.
func (d *Delivery) Source() *stomp.Message {
	return d.source
}

// Ack acknowledges the message delivery. Returns ErrDeliveryAlreadyHandled if the delivery has already been processed.
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

// Nack negatively acknowledges the message delivery, causing it to be requeued. Returns ErrDeliveryAlreadyHandled if the delivery has already been processed.
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

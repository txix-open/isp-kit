package consumer

import (
	"context"
	"sync"

	"github.com/go-stomp/stomp/v3"
	"github.com/pkg/errors"
)

type Consumer struct {
	Config
	conn          *stomp.Conn
	sub           *stomp.Subscription
	deliveryWg    *sync.WaitGroup
	workersWg     *sync.WaitGroup
	unexpectedErr chan error
	closeConsumer chan struct{}
}

func New(config Config) (*Consumer, error) {
	conn, err := stomp.Dial("tcp", config.Address, config.ConnOpts...)
	if err != nil {
		return nil, errors.WithMessagef(err, "stomp dial to %s", config.Address)
	}

	sub, err := conn.Subscribe(config.Queue, stomp.AckClientIndividual, config.SubscriptionOpts...)
	if err != nil {
		return nil, errors.WithMessagef(err, "stomp subscribe to %s", config.Queue)
	}

	return &Consumer{
		Config:        config,
		conn:          conn,
		sub:           sub,
		deliveryWg:    &sync.WaitGroup{},
		workersWg:     &sync.WaitGroup{},
		unexpectedErr: make(chan error, config.Concurrency),
		closeConsumer: make(chan struct{}),
	}, nil
}

func (c *Consumer) Run() error {
	for range c.Concurrency {
		c.workersWg.Add(1)
		go c.runWorker()
	}

	c.Observer.BeginConsuming(c)

	select {
	case err := <-c.unexpectedErr:
		c.workersWg.Wait()
		return err
	case <-c.closeConsumer:
		return nil
	}
}

func (c *Consumer) runWorker() {
	defer c.workersWg.Done()

	for {
		msg, err := c.sub.Read()
		if errors.Is(err, stomp.ErrCompletedSubscription) {
			return
		}
		if err != nil {
			c.Observer.Error(c, err)
			c.unexpectedErr <- err
			return
		}

		c.deliveryWg.Add(1)
		delivery := NewDelivery(c.deliveryWg, c.conn, msg)
		c.handler.Handle(context.Background(), delivery)
	}
}

func (c *Consumer) Close() error {
	defer func() {
		c.Observer.CloseDone(c)
		close(c.closeConsumer)
	}()

	c.Observer.CloseStart(c)

	err := c.sub.Unsubscribe()
	if err != nil {
		return errors.WithMessage(err, "stomp unsubscribe")
	}

	c.workersWg.Wait()
	c.deliveryWg.Wait()

	err = c.conn.Disconnect()
	if err != nil {
		return errors.WithMessage(err, "stomp disconnect")
	}

	return nil
}

package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
)

type Handler interface {
	Handle(ctx context.Context, delivery *Delivery)
}

type HandlerFunc func(ctx context.Context, delivery *Delivery)

func (f HandlerFunc) Handle(ctx context.Context, delivery *Delivery) {
	f(ctx, delivery)
}

type Middleware func(next Handler) Handler

type ConnOption = func(*stomp.Conn) error

type SubscriptionOption = func(*frame.Frame) error

type Config struct {
	Address          string
	Queue            string
	ConnOpts         []ConnOption
	Concurrency      int
	Middlewares      []Middleware
	SubscriptionOpts []SubscriptionOption
	Observer         Observer
	ReconnectTimeout time.Duration

	handler Handler
}

func NewConfig(address string, queue string, handler Handler, opts ...Option) Config {
	c := &Config{
		Address:          address,
		Queue:            queue,
		Concurrency:      1,
		ReconnectTimeout: 1 * time.Second,
		Observer:         NoopObserver{},
	}
	for _, opt := range opts {
		opt(c)
	}

	for i := len(c.Middlewares) - 1; i >= 0; i-- {
		handler = c.Middlewares[i](handler)
	}
	c.handler = handler

	return *c
}

func (c Config) String() string {
	return fmt.Sprintf("%s/%s", c.Address, c.Queue)
}

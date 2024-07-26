package consumer

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/handler"
	"github.com/txix-open/isp-kit/log"
	"go.uber.org/atomic"
)

type Middleware func(next Handler) Handler

type Handler interface {
	Handle(ctx context.Context, msg *kafka.Message) handler.Result
}

type HandlerFunc func(ctx context.Context, msg *kafka.Message) handler.Result

func (f HandlerFunc) Handle(ctx context.Context, msg *kafka.Message) handler.Result {
	return f(ctx, msg)
}

type Consumer struct {
	TopicName   string
	Middlewares []Middleware

	logger   log.Logger
	reader   *kafka.Reader
	handler  Handler
	wg       *sync.WaitGroup
	close    chan struct{}
	alive    *atomic.Bool
	observer Observer
}

func New(logger log.Logger, reader *kafka.Reader, handler Handler, opts ...Option) *Consumer {
	c := &Consumer{
		TopicName: reader.Config().Topic,
		reader:    reader,
		handler:   handler,
		wg:        &sync.WaitGroup{},
		logger:    logger,
		close:     make(chan struct{}),
		alive:     atomic.NewBool(true),
	}

	for _, opt := range opts {
		opt(c)
	}

	for i := len(c.Middlewares) - 1; i >= 0; i-- {
		handler = c.Middlewares[i](handler)
	}
	c.handler = handler

	return c
}

func (c *Consumer) Run(ctx context.Context) {
	c.wg.Add(1)
	go c.handleMessages(ctx)
}

func (c *Consumer) Close() error {
	close(c.close)
	err := c.reader.Close()
	if err != nil {
		return errors.WithMessage(err, "close kafka/Reader")
	}
	c.wg.Wait()
	c.alive.Store(false)
	return nil
}

func (c *Consumer) Healthcheck(ctx context.Context) error {
	if c.alive.Load() {
		return nil
	}
	return errors.New("could not fetch messages")
}

func (c *Consumer) handleMessages(ctx context.Context) {
	defer c.wg.Done()
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			c.alive.Store(false)
			c.observer.ConsumerError(*c, err)
			c.logger.Error(ctx, "kafka consumer: unexpected error during fetching messages", log.Any("error", err))

			select {
			case <-ctx.Done():
			case <-time.After(1 * time.Second):
			}

			continue
		}

		c.alive.Store(true)

		c.handleMessage(ctx, &msg)
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg *kafka.Message) {
	for {
		result := c.handler.Handle(ctx, msg)
		switch {
		case result.Commit:
			err := c.reader.CommitMessages(ctx, *msg)
			if err != nil {
				c.logger.Error(ctx, "kafka consumer: unexpected error during committing messages", log.Any("error", err))
			}
			return
		case result.Retry:
			select {
			case <-ctx.Done():
				return
			case <-c.close:
				return
			case <-time.After(result.RetryAfter):
				continue
			}
		}
	}
}

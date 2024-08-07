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
	Concurrency int

	logger     log.Logger
	reader     *kafka.Reader
	handler    Handler
	observer   Observer
	deliveryWg *sync.WaitGroup // ждет обработки всех delivery
	workersWg  *sync.WaitGroup // ждет завершения работы воркеров (<-c.workersStop / close(c.deliveries))
	deliveries chan Delivery   // канал доставок
	alive      *atomic.Bool
}

func New(logger log.Logger, reader *kafka.Reader, handler Handler, concurrency int, opts ...Option) *Consumer {
	if concurrency <= 0 {
		concurrency = 1
	}

	c := &Consumer{
		TopicName:   reader.Config().Topic,
		Concurrency: concurrency,
		reader:      reader,
		handler:     handler,
		logger:      logger,
		deliveryWg:  &sync.WaitGroup{},
		workersWg:   &sync.WaitGroup{},
		deliveries:  make(chan Delivery),
		alive:       atomic.NewBool(true),
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
	c.observer.BeginConsuming()
	go c.run(ctx)
}

func (c *Consumer) run(ctx context.Context) {
	for i := 0; i < c.Concurrency; i++ {
		c.workersWg.Add(1)
		go c.runWorker(ctx, i)
	}

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			c.alive.Store(false)
			c.observer.ConsumerError(err)

			select {
			case <-ctx.Done():
			case <-time.After(1 * time.Second):
			}

			continue
		}

		c.alive.Store(true)
		delivery := NewDelivery(c.deliveryWg, c.reader, &msg)
		c.deliveries <- *delivery
	}
}

//nolint:gosimple
func (c *Consumer) runWorker(ctx context.Context, num int) {
	defer c.workersWg.Done()

	for {
		select {
		case delivery, isOpen := <-c.deliveries:
			if !isOpen { //normal close
				return
			}
			c.deliveryWg.Add(1)
			c.handleMessage(ctx, &delivery)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, delivery *Delivery) {
	for {
		result := c.handler.Handle(ctx, delivery.Source())
		switch {
		case result.Commit:
			err := delivery.Commit(ctx)
			if err != nil {
				c.observer.ConsumerError(err)
			}
			return
		case result.Retry:
			select {
			case <-ctx.Done():
				delivery.donner.Done()
				return
			case <-time.After(result.RetryAfter):
				continue
			}
		}
	}
}

func (c *Consumer) Close() error {
	defer func() {
		c.deliveryWg.Wait()
		close(c.deliveries)
		c.workersWg.Wait()
		c.alive.Store(false)

		c.observer.CloseDone()
	}()
	c.observer.CloseStart()

	err := c.reader.Close()
	if err != nil {
		return errors.WithMessage(err, "close kafka.Reader")
	}

	return nil
}

func (c *Consumer) Healthcheck(ctx context.Context) error {
	if c.alive.Load() {
		return nil
	}
	return errors.New("could not fetch messages")
}

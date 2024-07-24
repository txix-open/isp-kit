package kafkax

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax/handler"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"go.uber.org/atomic"
)

type Consumer struct {
	reader    *kafka.Reader
	handler   handler.SyncHandlerAdapter
	wg        *sync.WaitGroup
	close     chan struct{}
	logger    log.Logger
	alive     *atomic.Bool
	TopicName string
	observer  Observer
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
		ctx := log.ToContext(ctx, log.String("requestId", requestid.Next()))
		c.logger.Debug(
			ctx,
			"kafka consumer: consume message",
			log.String("topic", msg.Topic),
			log.Int("partition", msg.Partition),
			log.Int64("offset", msg.Offset),
			log.ByteString("messageKey", msg.Key),
			log.ByteString("messageValue", msg.Value),
		)

		c.handleMessage(ctx, &msg)
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg *kafka.Message) {
	for {
		result := c.handler.Handle(ctx, msg)
		switch {
		case result.Commit:
			c.logger.Debug(ctx, "kafka consumer: message will be committed")
			err := c.reader.CommitMessages(ctx, *msg)
			if err != nil {
				c.logger.Error(ctx, "kafka consumer: unexpected error during committing messages", log.Any("error", err))
			}
			return
		case result.Retry:
			c.logger.Error(ctx, "kafka consumer: message will be retried",
				log.Any("error", result.RetryError), log.String("retryAfter", result.RetryAfter.String()),
			)
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

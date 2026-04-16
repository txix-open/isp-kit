// Package handler provides synchronous message processing with support for
// retry logic, commit handling, and middleware chains for cross-cutting
// concerns like logging, metrics, and panic recovery.
package handler

import (
	"context"
	"time"

	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/log"
)

// SyncHandlerAdapter defines the interface for synchronous message processing.
// Implementations should return a Result indicating whether to commit, retry,
// or skip the message.
type SyncHandlerAdapter interface {
	Handle(ctx context.Context, delivery *consumer.Delivery) Result
}

// Middleware is a function that wraps a SyncHandlerAdapter to add
// cross-cutting functionality such as logging, metrics, or panic recovery.
type Middleware func(next SyncHandlerAdapter) SyncHandlerAdapter

// SyncHandlerAdapterFunc is an adapter that allows a function to be used as
// a SyncHandlerAdapter.
type SyncHandlerAdapterFunc func(ctx context.Context, delivery *consumer.Delivery) Result

// Handle implements the SyncHandlerAdapter interface by calling the underlying
// function.
func (a SyncHandlerAdapterFunc) Handle(ctx context.Context, delivery *consumer.Delivery) Result {
	return a(ctx, delivery)
}

// Sync wraps a SyncHandlerAdapter with middleware support and handles the
// message lifecycle including committing, retrying, or skipping based on the
// returned Result.
type Sync struct {
	logger  log.Logger
	handler SyncHandlerAdapter
}

// NewSync creates a new Sync instance with the provided logger, adapter, and
// middlewares. Middlewares are applied in the order they are provided.
func NewSync(logger log.Logger, adapter SyncHandlerAdapter, middlewares ...Middleware) Sync {
	s := Sync{
		logger: logger,
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		adapter = middlewares[i](adapter)
	}
	s.handler = adapter
	return s
}

// Handle processes a message by calling the configured handler and acting
// on the returned Result. It supports automatic retry with backoff and
// error logging during offset committing.
func (r Sync) Handle(ctx context.Context, delivery *consumer.Delivery) {
	for {
		result := r.handler.Handle(ctx, delivery)
		switch {
		case result.Commit:
			err := delivery.Commit(ctx)
			if err != nil {
				r.logger.Error(
					ctx, "kafka consumer: unexpected error during committing message",
					log.Any("error", err),
				)
			}
			return
		case result.Retry:
			select {
			case <-ctx.Done():
				delivery.Done()
				return
			case <-time.After(result.RetryAfter):
				continue
			}
		default:
			delivery.Done()
			return
		}
	}
}

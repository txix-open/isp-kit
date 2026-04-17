package handler

import (
	"context"

	"github.com/go-stomp/stomp/v3"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/panic_recovery"
)

// Middleware is a function that wraps a HandlerAdapter with additional functionality.
type Middleware func(next HandlerAdapter) HandlerAdapter

// Log creates a middleware that logs message processing results.
func Log(logger log.Logger) Middleware {
	return func(next HandlerAdapter) HandlerAdapter {
		return AdapterFunc(func(ctx context.Context, msg *stomp.Message) Result {
			destination := msg.Destination

			result := next.Handle(ctx, msg)

			switch {
			case result.Ack:
				logger.Debug(
					ctx,
					"stomp client: message will be acknowledged",
					log.String("destination", destination),
				)
			case result.Requeue && result.Err != nil:
				logger.Error(
					ctx,
					"stomp client: message will be requeued",
					log.Any("error", result.Err),
					log.String("destination", destination),
				)
			case result.Requeue:
				logger.Debug(
					ctx,
					"stomp client: message will be requeued",
					log.String("destination", destination),
				)
			}

			return result
		})
	}
}

// Recovery creates a middleware that recovers from panics and requeues the message.
func Recovery() Middleware {
	return func(next HandlerAdapter) HandlerAdapter {
		return AdapterFunc(func(ctx context.Context, msg *stomp.Message) (res Result) {
			defer panic_recovery.Recover(func(err error) {
				res.Err = err
				res.Requeue = true
			})
			return next.Handle(ctx, msg)
		})
	}
}

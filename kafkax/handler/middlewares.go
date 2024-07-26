package handler

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/log"
)

func Log(logger log.Logger) Middleware {
	return func(next SyncHandlerAdapter) SyncHandlerAdapter {
		return SyncHandlerAdapterFunc(func(ctx context.Context, msg *kafka.Message) Result {
			partition := msg.Partition
			offset := msg.Offset

			result := next.Handle(ctx, msg)

			switch {
			case result.Commit:
				logger.Debug(
					ctx,
					"kafka client: message will be commited",
					log.Int("partition", partition),
					log.Int64("offset", offset),
				)
			case result.Retry:
				logger.Error(
					ctx,
					"kafka client: message will be retried",
					log.Int("partition", partition),
					log.Int64("offset", offset),
					log.Any("error", result.RetryError),
					log.String("retryAfter", result.RetryAfter.String()),
				)
			}

			return result
		})
	}
}

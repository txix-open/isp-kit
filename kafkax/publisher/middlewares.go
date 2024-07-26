package publisher

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/protocol"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

const RequestIdHeader = "x-request-id"

type PublisherMetricStorage interface {
	ObservePublishDuration(topic string, partition int, offset int64, t time.Duration)
	ObservePublishMsgSize(topic string, partition int, offset int64, size int)
	IncPublishError(topic string, partition int, offset int64)
}

func PublisherMetrics(storage PublisherMetricStorage) Middleware {
	return func(next RoundTripper) RoundTripper {
		return RoundTripperFunc(func(ctx context.Context, msg *kafka.Message) error {
			storage.ObservePublishMsgSize(msg.Topic, msg.Partition, msg.Offset, len(msg.Value))
			start := time.Now()

			err := next.Publish(ctx, msg)
			if err != nil {
				storage.IncPublishError(msg.Topic, msg.Partition, msg.Offset)
			}

			storage.ObservePublishDuration(msg.Topic, msg.Partition, msg.Offset, time.Since(start))
			return err
		})
	}
}

func PublisherLog(logger log.Logger) Middleware {
	return func(next RoundTripper) RoundTripper {
		return RoundTripperFunc(func(ctx context.Context, msg *kafka.Message) error {
			logger.Debug(
				ctx,
				"kafka client: publish message",
				log.String("topic", msg.Topic),
				log.Int("partition", msg.Partition),
				log.ByteString("messageKey", msg.Key),
				log.ByteString("messageValue", msg.Value),
			)

			return next.Publish(ctx, msg)
		})
	}
}

func PublisherRequestId() Middleware {
	return func(next RoundTripper) RoundTripper {
		return RoundTripperFunc(func(ctx context.Context, msg *kafka.Message) error {
			requestId := requestid.FromContext(ctx)

			if requestId == "" {
				requestId = requestid.Next()
			}

			msg.Headers = append(msg.Headers, protocol.Header{
				Key:   RequestIdHeader,
				Value: []byte(requestId),
			})

			return next.Publish(ctx, msg)
		})
	}
}

package kafkax

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/protocol"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/handler"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

const RequestIdHeader = "x-request-id"

type PublisherMetricStorage interface {
	ObservePublishDuration(topic string, partition int, offset int64, t time.Duration)
	ObservePublishMsgSize(topic string, partition int, offset int64, size int)
	IncPublishError(topic string, partition int, offset int64)
}

func PublisherMetrics(storage PublisherMetricStorage) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, msg *kafka.Message) error {
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

func PublisherLog(logger log.Logger) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, msg *kafka.Message) error {
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

func PublisherRequestId() publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, msg *kafka.Message) error {
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

type ConsumerMetricStorage interface {
	ObserveConsumeDuration(topic string, partition int, offset int64, t time.Duration)
	ObserveConsumeMsgSize(topic string, partition int, offset int64, size int)
	IncCommitCount(topic string, partition int, offset int64)
	IncRetryCount(topic string, partition int, offset int64)
}

func ConsumerMetrics(metricStorage ConsumerMetricStorage) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, msg *kafka.Message) handler.Result {
			topic := msg.Topic
			partition := msg.Partition
			offset := msg.Offset
			start := time.Now()

			result := next.Handle(ctx, msg)

			metricStorage.ObserveConsumeDuration(topic, partition, offset, time.Since(start))
			metricStorage.ObserveConsumeMsgSize(topic, partition, offset, len(msg.Value))

			switch {
			case result.Commit:
				metricStorage.IncCommitCount(topic, partition, offset)
			case result.Retry:
				metricStorage.IncRetryCount(topic, partition, offset)
			}

			return result
		})
	}
}

func ConsumerLog(logger log.Logger) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, msg *kafka.Message) handler.Result {
			logger.Debug(
				ctx,
				"kafka consumer: consume message",
				log.String("topic", msg.Topic),
				log.Int("partition", msg.Partition),
				log.Int64("offset", msg.Offset),
				log.ByteString("messageKey", msg.Key),
				log.ByteString("messageValue", msg.Value),
			)
			return next.Handle(ctx, msg)
		})
	}
}

func ConsumerRequestId() consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, msg *kafka.Message) handler.Result {
			requestId := ""

			if msg.Headers != nil {
				requestId = Get(msg.Headers, RequestIdHeader)
			}

			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String("requestId", requestId))

			return next.Handle(ctx, msg)
		})
	}
}

func Get(headers []kafka.Header, key string) string {
	for _, header := range headers {
		if header.Key == key {
			return string(header.Value)
		}
	}

	return ""
}

package kafkax

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/protocol"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

const RequestIdHeader = "x-request-id"

type PublisherMetricStorage interface {
	ObservePublishDuration(topic string, t time.Duration)
	ObservePublishMsgSize(topic string, size int)
	IncPublishError(topic string)
}

func PublisherMetrics(storage PublisherMetricStorage) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, msgs ...kafka.Message) error {
			topic := msgs[0].Topic

			for _, msg := range msgs {
				storage.ObservePublishMsgSize(msg.Topic, len(msg.Value))
			}
			start := time.Now()

			err := next.Publish(ctx, msgs...)
			if err != nil {
				storage.IncPublishError(topic)
			}

			storage.ObservePublishDuration(topic, time.Since(start))

			return err
		})
	}
}

func PublisherLog(logger log.Logger) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, msgs ...kafka.Message) error {
			for _, msg := range msgs {
				logger.Debug(
					ctx,
					"kafka client: publish message",
					log.String("topic", msg.Topic),
					log.Int("partition", msg.Partition),
					log.ByteString("messageKey", msg.Key),
					log.ByteString("messageValue", msg.Value),
				)
			}

			return next.Publish(ctx, msgs...)
		})
	}
}

func PublisherRequestId() publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, msgs ...kafka.Message) error {
			for i, msg := range msgs {
				requestId := requestid.FromContext(ctx)

				if requestId == "" {
					requestId = requestid.Next()
				}

				msgs[i].Headers = append(msg.Headers, protocol.Header{
					Key:   RequestIdHeader,
					Value: []byte(requestId),
				})
			}

			return next.Publish(ctx, msgs...)
		})
	}
}

func ConsumerLog(logger log.Logger) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			logger.Debug(
				ctx,
				"kafka consumer: consume message",
				log.String("topic", delivery.Source().Topic),
				log.Int("partition", delivery.Source().Partition),
				log.Int64("offset", delivery.Source().Offset),
				log.ByteString("messageKey", delivery.Source().Key),
				log.ByteString("messageValue", delivery.Source().Value),
			)
			next.Handle(ctx, delivery)
		})
	}
}

func ConsumerRequestId() consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			requestId := GetHeaderValue(delivery.Source().Headers, RequestIdHeader)

			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String("requestId", requestId))

			next.Handle(ctx, delivery)
		})
	}
}

func GetHeaderValue(headers []kafka.Header, key string) string {
	for _, header := range headers {
		if header.Key == key {
			return string(header.Value)
		}
	}

	return ""
}

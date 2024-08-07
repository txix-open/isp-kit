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
	ObservePublishDuration(requestId string, t time.Duration)
	ObservePublishMsgSize(topic string, size int)
	IncPublishError(requestId string)
}

func PublisherMetrics(storage PublisherMetricStorage) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, msgs ...kafka.Message) error {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				// requestId устанавливается в PublisherRequestId(), здесь на случай отсутствия той middleware
				requestId = requestid.Next()
			}

			for _, msg := range msgs {
				storage.ObservePublishMsgSize(msg.Topic, len(msg.Value))
			}
			start := time.Now()

			err := next.Publish(ctx, msgs...)
			if err != nil {
				storage.IncPublishError(requestId)
			}

			storage.ObservePublishDuration(requestId, time.Since(start))

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

type ConsumerMetricStorage interface {
	ObserveConsumeDuration(topic string, t time.Duration)
	ObserveConsumeMsgSize(topic string, size int)
	IncCommitCount(topic string)
	IncRetryCount(topic string)
}

func ConsumerMetrics(metricStorage ConsumerMetricStorage) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, msg *kafka.Message) handler.Result {
			topic := msg.Topic
			start := time.Now()

			result := next.Handle(ctx, msg)

			metricStorage.ObserveConsumeDuration(topic, time.Since(start))
			metricStorage.ObserveConsumeMsgSize(topic, len(msg.Value))

			switch {
			case result.Commit:
				metricStorage.IncCommitCount(topic)
			case result.Retry:
				metricStorage.IncRetryCount(topic)
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
				requestId = GetHeaderValue(msg.Headers, RequestIdHeader)
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

func GetHeaderValue(headers []kafka.Header, key string) string {
	for _, header := range headers {
		if header.Key == key {
			return string(header.Value)
		}
	}

	return ""
}

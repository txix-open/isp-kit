package kafkax

import (
	"context"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

// PublisherMetricStorage defines the interface for publisher metrics storage.
type PublisherMetricStorage interface {
	ObservePublishDuration(topic string, t time.Duration)
	ObservePublishMsgSize(topic string, size int)
	IncPublishError(topic string)
}

// PublisherMetrics creates a middleware that records metrics for message
// publishing operations, including duration, message size, and error counts.
func PublisherMetrics(storage PublisherMetricStorage) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, rs ...*kgo.Record) error {
			if len(rs) < 1 {
				return nil
			}
			topic := rs[0].Topic

			for _, r := range rs {
				storage.ObservePublishMsgSize(r.Topic, len(r.Value))
			}
			start := time.Now()

			err := next.Publish(ctx, rs...)
			if err != nil {
				storage.IncPublishError(topic)
			}

			storage.ObservePublishDuration(topic, time.Since(start))

			return err
		})
	}
}

// PublisherLog creates a middleware that logs publish operations. When logBody
// is true, the message body is included in the log output.
func PublisherLog(logger log.Logger, logBody bool) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, rs ...*kgo.Record) error {
			for _, r := range rs {
				fields := []log.Field{
					log.String("topic", r.Topic),
					log.Int32("partition", r.Partition),
					log.ByteString("messageKey", r.Key),
					log.Int("bodySize", len(r.Value)),
				}
				if logBody {
					fields = append(fields, log.ByteString("body", r.Value))
				}
				logger.Debug(
					ctx,
					"kafka client: publish message",
					fields...,
				)
			}

			return next.Publish(ctx, rs...)
		})
	}
}

// PublisherRequestId creates a middleware that propagates request IDs to Kafka
// message headers. If no request ID is present in the context, a new one is
// generated.
func PublisherRequestId() publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, msgs ...*kgo.Record) error {
			for i := range msgs {
				requestId := requestid.FromContext(ctx)

				if requestId == "" {
					requestId = requestid.Next()
				}

				msgs[i].Headers = append(msgs[i].Headers, kgo.RecordHeader{
					Key:   requestid.Header,
					Value: []byte(requestId),
				})
			}

			return next.Publish(ctx, msgs...)
		})
	}
}

// Retrier defines an interface for retry logic implementations.
type Retrier interface {
	Do(ctx context.Context, f func() error) error
}

// PublisherRetry creates a middleware for retrying message publications.
// It is recommended to use this middleware after logging,
// to avoid duplicate logging of publication attempts.
func PublisherRetry(retrier Retrier) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, rs ...*kgo.Record) error {
			return retrier.Do(ctx, func() error {
				return next.Publish(ctx, rs...)
			})
		})
	}
}

// ConsumerLog creates a middleware that logs consume operations. When logBody
// is true, the message body is included in the log output.
func ConsumerLog(logger log.Logger, logBody bool) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			fields := []log.Field{
				log.String("topic", delivery.Source().Topic),
				log.Int32("partition", delivery.Source().Partition),
				log.Int64("offset", delivery.Source().Offset),
				log.ByteString("messageKey", delivery.Source().Key),
				log.Int("bodySize", len(delivery.Source().Value)),
			}
			if logBody {
				fields = append(fields, log.ByteString("body", delivery.Source().Value))
			}
			logger.Debug(
				ctx,
				"kafka consumer: consume message",
				fields...,
			)
			next.Handle(ctx, delivery)
		})
	}
}

// ConsumerRequestId creates a middleware that extracts request IDs from Kafka
// message headers and adds them to the context. If no request ID is found,
// a new one is generated.
func ConsumerRequestId() consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			requestId := GetHeaderValue(delivery.Source().Headers, requestid.Header)

			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String(requestid.LogKey, requestId))

			next.Handle(ctx, delivery)
		})
	}
}

// GetHeaderValue retrieves the value of a header with the specified key from
// the provided headers slice. Returns an empty string if the key is not found.
func GetHeaderValue(headers []kgo.RecordHeader, key string) string {
	for _, header := range headers {
		if header.Key == key {
			return string(header.Value)
		}
	}

	return ""
}

package grmqx

import (
	"context"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
)

// PublisherLog creates a publisher middleware that logs published messages.
// When logBody is true, the message body is included in the log output.
func PublisherLog(logger log.Logger, logBody bool) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			fields := []log.Field{
				log.String("exchange", exchange),
				log.String("routingKey", routingKey),
				log.Int("bodySize", len(msg.Body)),
			}
			if logBody {
				fields = append(fields, log.ByteString("body", msg.Body))
			}
			logger.Debug(
				ctx,
				"rmq client: publish message",
				fields...,
			)
			return next.Publish(ctx, exchange, routingKey, msg)
		})
	}
}

// PublisherRequestId creates a publisher middleware that generates and injects request IDs
// into message headers. If a request ID already exists in the context, it is used; otherwise,
// a new one is generated.
func PublisherRequestId() publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}
			if msg.Headers == nil {
				msg.Headers = amqp091.Table{}
			}
			msg.Headers[requestid.Header] = requestId
			return next.Publish(ctx, exchange, routingKey, msg)
		})
	}
}

// Retrier defines an interface for retry logic.
type Retrier interface {
	Do(ctx context.Context, f func() error) error
}

// PublisherRetry creates a middleware for retrying message publications.
//
// Place this middleware after logging and encoding middlewares so that
// publication attempts are logged only once and payloads are encoded only once.
func PublisherRetry(retrier Retrier) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			return retrier.Do(ctx, func() error {
				return next.Publish(ctx, exchange, routingKey, msg)
			})
		})
	}
}

// PublisherMetricStorage defines an interface for publisher metrics storage.
type PublisherMetricStorage interface {
	ObservePublishDuration(exchange string, routingKey string, t time.Duration)
	ObservePublishMsgSize(exchange string, routingKey string, size int)
	IncPublishError(exchange string, routingKey string)
}

// PublisherMetrics creates a middleware that collects publisher metrics including
// message size, publication duration, and error counts.
func PublisherMetrics(storage PublisherMetricStorage) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			start := time.Now()
			err := next.Publish(ctx, exchange, routingKey, msg)
			if err != nil {
				storage.IncPublishError(exchange, routingKey)
			}
			storage.ObservePublishDuration(exchange, routingKey, time.Since(start))
			storage.ObservePublishMsgSize(exchange, routingKey, len(msg.Body))
			return err
		})
	}
}

type EncoderCodec interface {
	Type() string
	EncodeBytes(body []byte) ([]byte, error)
}

// EncodeMessage creates a publisher middleware that transparently handles message encoding.
//
// If message body size > encodeThreshold, message body is encoded using codec.
//
// Important:
//   - Should be placed after logging middleware if logs must see plain data.
func EncodeMessage(codec EncoderCodec, encodeThresholdBytes int) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, exchange string, routingKey string, msg *amqp091.Publishing) error {
			if len(msg.Body) > encodeThresholdBytes {
				encoded, err := codec.EncodeBytes(msg.Body)
				if err != nil {
					return errors.WithMessagef(err, "encode by '%s' encoder", codec.Type())
				}
				msg.ContentEncoding = codec.Type()
				msg.Body = encoded
			}
			return next.Publish(ctx, exchange, routingKey, msg)
		})
	}
}

// ConsumerLog creates a consumer middleware that logs consumed messages.
// When logBody is true, the message body is included in the log output.
func ConsumerLog(logger log.Logger, logBody bool) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			fields := []log.Field{
				log.String("exchange", delivery.Source().Exchange),
				log.String("routingKey", delivery.Source().RoutingKey),
				log.Int("bodySize", len(delivery.Source().Body)),
			}
			if logBody {
				fields = append(fields, log.ByteString("body", delivery.Body))
			}
			logger.Debug(
				ctx,
				"rmq client: consume message",
				fields...,
			)
			next.Handle(ctx, delivery)
		})
	}
}

// ConsumerRequestId creates a consumer middleware that extracts request IDs from message headers
// and propagates them in the context. If no request ID is found, a new one is generated.
// It also adds the request ID to the context logger.
func ConsumerRequestId() consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			requestId := ""
			headers := delivery.Source().Headers
			if headers != nil {
				value, ok := headers[requestid.Header].(string)
				if ok {
					requestId = value
				}
			}
			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String(requestid.LogKey, requestId))
			next.Handle(ctx, delivery)
		})
	}
}

type DecodeCodec interface {
	Type() string
	DecodeBytes(body []byte) ([]byte, error)
}

// DecodeMessage creates a consumer middleware that transparently handles message encoding.
//
// If contentEncoding matches codec, message body is decoded.
//
// Important:
//   - Should be placed before logging middleware if logs must see decoded data.
func DecodeMessage(codec DecodeCodec, logger log.Logger) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			if delivery.Source().ContentEncoding != codec.Type() {
				next.Handle(ctx, delivery)
				return
			}

			decoded, err := codec.DecodeBytes(delivery.Body)
			if err != nil {
				logger.Error(ctx, "failed to decode message",
					log.String("contentEncoding", delivery.Source().ContentEncoding),
					log.Any("error", err),
				)
				err := delivery.Nack(false)
				if err != nil {
					logger.Error(ctx, "rmq client: nack message error", log.Any("error", err))
				}
				return
			}

			delivery.Body = decoded
			next.Handle(ctx, delivery)
		})
	}
}

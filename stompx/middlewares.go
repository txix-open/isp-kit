package stompx

import (
	"context"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/publisher"
)

func PublisherPersistent() publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, queue string, msg *publisher.Message) error {
			msg = msg.WithHeader("persistent", "true")
			return next.Publish(ctx, queue, msg)
		})
	}
}

func PublisherLog(logger log.Logger, logBody bool) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, queue string, msg *publisher.Message) error {
			fields := []log.Field{
				log.String("queue", queue),
			}
			if logBody {
				fields = append(fields, log.ByteString("body", msg.Body))
			}
			logger.Debug(
				ctx,
				"stomp client: publish message",
				fields...,
			)
			return next.Publish(ctx, queue, msg)
		})
	}
}

func PublisherRequestId() publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, queue string, msg *publisher.Message) error {
			requestId := requestid.FromContext(ctx)
			if requestId == "" {
				requestId = requestid.Next()
			}
			msg = msg.WithHeader(requestid.Header, requestId)
			return next.Publish(ctx, queue, msg)
		})
	}
}

type Retrier interface {
	Do(ctx context.Context, f func() error) error
}

func PublisherRetry(retrier Retrier) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, queue string, msg *publisher.Message) error {
			return retrier.Do(ctx, func() error {
				return next.Publish(ctx, queue, msg)
			})
		})
	}
}

func ConsumerLog(logger log.Logger, logBody bool) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			fields := []log.Field{
				log.String("queue", delivery.Source().Destination),
			}
			if logBody {
				fields = append(fields, log.ByteString("body", delivery.Source().Body))
			}
			logger.Debug(
				ctx,
				"stomp client: consume message",
				fields...,
			)
			next.Handle(ctx, delivery)
		})
	}
}

func ConsumerRequestId() consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			requestId := ""
			headers := delivery.Source().Header
			if headers != nil {
				requestId = headers.Get(requestid.Header)
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

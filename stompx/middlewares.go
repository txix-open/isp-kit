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

func PublisherLog(logger log.Logger) publisher.Middleware {
	return func(next publisher.RoundTripper) publisher.RoundTripper {
		return publisher.RoundTripperFunc(func(ctx context.Context, queue string, msg *publisher.Message) error {
			logger.Debug(
				ctx,
				"stomp client: publish message",
				log.String("queue", queue),
				log.ByteString("body", msg.Body),
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
			msg = msg.WithHeader(requestid.RequestIdHeader, requestId)
			return next.Publish(ctx, queue, msg)
		})
	}
}

func ConsumerLog(logger log.Logger) consumer.Middleware {
	return func(next consumer.Handler) consumer.Handler {
		return consumer.HandlerFunc(func(ctx context.Context, delivery *consumer.Delivery) {
			logger.Debug(
				ctx,
				"stomp client: consume message",
				log.String("queue", delivery.Source().Destination),
				log.ByteString("body", delivery.Source().Body),
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
				requestId = headers.Get(requestid.RequestIdHeader)
			}
			if requestId == "" {
				requestId = requestid.Next()
			}
			ctx = requestid.ToContext(ctx, requestId)
			ctx = log.ToContext(ctx, log.String("requestId", requestId))
			next.Handle(ctx, delivery)
		})
	}
}

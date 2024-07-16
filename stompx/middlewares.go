package stompx

import (
	"context"
	"github.com/go-stomp/stomp/v3"

	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/publisher"
)

const (
	RequestIdHeader = "x-request-id"
)

type Middleware func(next HandlerAdapter) HandlerAdapter

func Log(logger log.Logger) Middleware {
	return func(next HandlerAdapter) HandlerAdapter {
		return AdapterFunc(func(ctx context.Context, msg *stomp.Message) Result {
			destination := msg.Destination

			result := next.Handle(ctx, msg)

			switch {
			case result.ack:
				logger.Debug(
					ctx,
					"stomp client: message will be acknowledged",
					log.String("destination", destination),
				)
			case result.requeue:
				logger.Error(
					ctx,
					"stomp client: message will be requeued",
					log.Any("error", result.err),
					log.String("destination", destination),
				)
			}

			return result
		})
	}
}

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
			msg = msg.WithHeader(RequestIdHeader, requestId)
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
				requestId = headers.Get(RequestIdHeader)
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

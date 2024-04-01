package tests

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/grmqx/handler"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/grmqt"
)

func TestRequestIdChain(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	expectedRequestId := requestid.Next()

	pubCfg1 := grmqx.Publisher{
		Exchange:   "",
		RoutingKey: "queue1",
	}
	pub1 := pubCfg1.DefaultPublisher(grmqx.PublisherLog(test.Logger()))
	consumerCfg1 := grmqx.Consumer{
		Queue: "queue1",
	}
	pubCfg2 := grmqx.Publisher{
		RoutingKey: "queue2",
	}
	pub2 := pubCfg2.DefaultPublisher(grmqx.PublisherLog(test.Logger()))
	consumerCfg2 := grmqx.Consumer{
		Queue: "queue2",
	}

	handler1 := grmqx.NewResultHandler(
		test.Logger(),
		handler.SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) handler.Result {
			err := pub2.Publish(ctx, &amqp091.Publishing{})
			require.NoError(err)
			return handler.Ack()
		}),
	)
	consumer1 := consumerCfg1.DefaultConsumer(handler1, grmqx.ConsumerLog(test.Logger()))

	await := make(chan struct{})
	handler2 := grmqx.NewResultHandler(
		test.Logger(),
		handler.SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) handler.Result {
			requestId := requestid.FromContext(ctx)
			require.EqualValues(expectedRequestId, requestId)
			close(await)
			return handler.Ack()
		}),
	)
	consumer2 := consumerCfg2.DefaultConsumer(handler2, grmqx.ConsumerLog(test.Logger()))

	testCli := grmqt.New(test)
	cli := grmqx.New(test.Logger())
	t.Cleanup(func() {
		cli.Close()
	})
	cfg := grmqx.NewConfig(
		testCli.ConnectionConfig().Url(),
		grmqx.WithPublishers(pub1, pub2),
		grmqx.WithConsumers(consumer1, consumer2),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg1, consumerCfg2)),
	)
	err := cli.Upgrade(context.Background(), cfg)
	require.NoError(err)

	ctx := requestid.ToContext(context.Background(), expectedRequestId)
	ctx = log.ToContext(ctx, log.String("requestId", expectedRequestId))
	err = pub1.Publish(ctx, &amqp091.Publishing{})
	require.NoError(err)

	select {
	case <-await:
	case <-time.After(5 * time.Second):
		require.Fail("handler wasn't called")
	}
}

func TestRetry(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher()
	callCount := atomic.Int32{}
	handler := grmqx.NewResultHandler(
		test.Logger(),
		handler.SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) handler.Result {
			callCount.Add(1)
			return handler.Retry(errors.New("some error"))
		}),
	)
	consumerCfg := grmqx.Consumer{
		Queue: "test",
		RetryPolicy: &grmqx.RetryPolicy{
			FinallyMoveToDlq: true,
			Retries: []grmqx.RetryConfig{{
				DelayInMs:   300,
				MaxAttempts: 3,
			}},
		},
	}
	consumer := consumerCfg.DefaultConsumer(handler, grmqx.ConsumerLog(test.Logger()))
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg)),
	)
	cli.Upgrade(config)

	err := pub.Publish(context.Background(), &amqp091.Publishing{})
	require.NoError(err)

	time.Sleep(2 * time.Second)

	require.EqualValues(4, callCount.Load())
	require.EqualValues(1, cli.QueueLength("test.DLQ"))
}

func TestBatchHandler(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher()
	deliveryCount := atomic.Int32{}
	handler := grmqx.BatchHandlerAdapterFunc(func(batch []grmqx.BatchItem) {
		for _, item := range batch {
			err := item.Delivery.Ack()
			require.NoError(err)
			deliveryCount.Add(1)
		}
	})
	consumerCfg := grmqx.BatchConsumer{
		Queue:             "test",
		BatchSize:         100,
		PurgeIntervalInMs: 60000,
	}
	consumer := consumerCfg.DefaultConsumer(handler, grmqx.ConsumerLog(test.Logger()))
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.ConsumerConfig())),
	)
	cli.Upgrade(config)

	for i := 0; i < 101; i++ {
		err := pub.Publish(context.Background(), &amqp091.Publishing{})
		require.NoError(err)
	}

	time.Sleep(2 * time.Second)

	cli.GrmqxCli.Close()

	require.EqualValues(101, deliveryCount.Load())
	require.EqualValues(0, cli.QueueLength("test"))
}

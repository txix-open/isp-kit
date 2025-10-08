package grmqx_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/grmqx/batch_handler"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/grmqt"
)

func TestBatchHandler(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher()
	deliveryCount := atomic.Int32{}

	handler := grmqx.NewResultBatchHandler(
		test.Logger(),
		batch_handler.SyncHandlerAdapterFunc(func(batch batch_handler.BatchItems) {
			for _, item := range batch {
				item.Ack()
				deliveryCount.Add(1)
			}
		}),
	)
	consumerCfg := grmqx.BatchConsumer{
		Queue:             "test",
		BatchSize:         100,
		PurgeIntervalInMs: 60000,
	}
	consumer := consumerCfg.DefaultConsumer(handler, grmqx.ConsumerLog(test.Logger(), true))
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.ConsumerConfig())),
	)
	cli.Upgrade(config)

	for range 101 {
		err := pub.Publish(t.Context(), &amqp091.Publishing{})
		require.NoError(err)
	}

	time.Sleep(2 * time.Second)

	cli.GrmqxCli.Close()

	require.EqualValues(101, deliveryCount.Load())
	require.EqualValues(0, cli.QueueLength("test"))
}

//nolint:gosec
func TestBatchHandlerRetry(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher()
	callCount := atomic.Int32{}

	handler := grmqx.NewResultBatchHandler(
		test.Logger(),
		batch_handler.SyncHandlerAdapterFunc(func(batch batch_handler.BatchItems) {
			batch.RetryAll(errors.New("test error"))
			callCount.Add(int32(len(batch)))
		}),
	)
	consumerCfg := grmqx.BatchConsumer{
		Queue: "test",
		RetryPolicy: &grmqx.RetryPolicy{
			FinallyMoveToDlq: true,
			Retries: []grmqx.RetryConfig{{
				DelayInMs:   300,
				MaxAttempts: 3,
			}},
		},
		BatchSize:         10,
		PurgeIntervalInMs: 60000,
	}
	consumer := consumerCfg.DefaultConsumer(handler, grmqx.ConsumerLog(test.Logger(), true))
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.ConsumerConfig())),
	)
	cli.Upgrade(config)

	for range 10 {
		err := pub.Publish(t.Context(), &amqp091.Publishing{})
		require.NoError(err)
	}

	time.Sleep(2 * time.Second)

	cli.GrmqxCli.Close()

	require.EqualValues(40, callCount.Load())
	require.EqualValues(10, cli.QueueLength("test.DLQ"))
}

func TestBatchRecovery(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher()
	deliveryCount := atomic.Int32{}

	handler := grmqx.NewResultBatchHandler(
		test.Logger(),
		batch_handler.SyncHandlerAdapterFunc(func(batch batch_handler.BatchItems) {
			for _, item := range batch {
				if deliveryCount.Load() == 5 {
					panic(errors.New("test panic"))
				}

				item.Ack()
				deliveryCount.Add(1)
			}
		}),
	)
	consumerCfg := grmqx.BatchConsumer{
		Queue:             "test",
		Dlq:               true,
		BatchSize:         10,
		PurgeIntervalInMs: 60000,
	}
	consumer := consumerCfg.DefaultConsumer(handler, grmqx.ConsumerLog(test.Logger(), true))
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.ConsumerConfig())),
	)
	cli.Upgrade(config)

	for range 10 {
		err := pub.Publish(t.Context(), &amqp091.Publishing{})
		require.NoError(err)
	}

	time.Sleep(1 * time.Second)

	cli.GrmqxCli.Close()

	require.EqualValues(5, deliveryCount.Load())
	require.EqualValues(0, cli.QueueLength("test"))
	require.EqualValues(5, cli.QueueLength("test.DLQ"))
}

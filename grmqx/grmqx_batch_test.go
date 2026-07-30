//nolint:gosec
package grmqx_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/isp-kit/codec"
	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/grmqx/batch_handler"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/fake"
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
			deliveryCount.Add(int32(len(batch)))
			batch.AckAll()
		}),
	)
	consumerCfg := grmqx.BatchConsumer{
		Queue:             "test",
		BatchSize:         100,
		PurgeIntervalInMs: 200,
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

	require.Eventually(func() bool {
		return deliveryCount.Load() == 101
	}, 2*time.Second, 100*time.Millisecond)

	require.EqualValues(101, deliveryCount.Load())
	require.EqualValues(0, cli.QueueLength("test"))
}

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

	require.EqualValues(5, deliveryCount.Load())
	require.EqualValues(0, cli.QueueLength("test"))
	require.EqualValues(5, cli.QueueLength("test.DLQ"))
}

func TestBatch_EncodedPublish_DecodeConsumer(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	codec := codec.Default
	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher(grmqx.PublisherLog(test.Logger(), true), grmqx.EncodeMessage(codec, 10))
	callCount := atomic.Int32{}

	publish := map[string]any{
		"field":  "value",
		"field2": fake.It[string](),
		"field3": fake.It[int](),
	}
	rawPublish, err := json.Marshal(publish)
	require.NoError(err)

	handler := grmqx.NewResultBatchHandler(
		test.Logger(),
		batch_handler.SyncHandlerAdapterFunc(func(batch batch_handler.BatchItems) {
			callCount.Add(int32(len(batch)))

			for _, item := range batch {
				require.Equal(codec.Type(), item.Delivery.Source().ContentEncoding)
				require.Equal(rawPublish, item.Delivery.Body)
			}
			batch.AckAll()
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
	consumer := consumerCfg.DefaultConsumer(
		handler,
		grmqx.DecodeMessage(codec, test.Logger()),
		grmqx.ConsumerLog(test.Logger(), true),
	)
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.ConsumerConfig())),
	)
	cli.Upgrade(config)

	for range 10 {
		err := pub.Publish(t.Context(), &amqp091.Publishing{Body: rawPublish})
		require.NoError(err)
	}

	require.Eventually(func() bool {
		return callCount.Load() == 10
	}, 2*time.Second, 100*time.Millisecond)

	require.EqualValues(10, callCount.Load())
	require.Empty(cli.QueueLength("test"))
	require.Empty(cli.QueueLength("test.DLQ"))
}
func TestBatch_EncodedPublish_DefaultConsumer_Retry(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	codec := codec.Default
	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher(grmqx.PublisherLog(test.Logger(), true), grmqx.EncodeMessage(codec, 10))
	callCount := atomic.Int32{}

	publish := map[string]any{
		"field":  "value",
		"field2": fake.It[string](),
		"field3": fake.It[int](),
	}
	rawPublish, err := json.Marshal(publish)
	require.NoError(err)

	expectedPublish, err := codec.EncodeBytes(rawPublish)
	require.NoError(err)

	handler := grmqx.NewResultBatchHandler(
		test.Logger(),
		batch_handler.SyncHandlerAdapterFunc(func(batch batch_handler.BatchItems) {
			callCount.Add(int32(len(batch)))

			for _, item := range batch {
				require.Equal(codec.Type(), item.Delivery.Source().ContentEncoding)
				require.Equal(expectedPublish, item.Delivery.Body)
			}
			batch.RetryAll(errors.Errorf("some error"))
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
		err := pub.Publish(t.Context(), &amqp091.Publishing{Body: rawPublish})
		require.NoError(err)
	}

	require.Eventually(func() bool {
		return callCount.Load() == 40
	}, 2*time.Second, 100*time.Millisecond)

	require.EqualValues(40, callCount.Load())
	require.Empty(cli.QueueLength("test"))
	require.EqualValues(10, cli.QueueLength("test.DLQ"))

	delivery := cli.DrainMessage("test.DLQ")
	require.Equal(codec.Type(), delivery.ContentEncoding)
	require.Equal(expectedPublish, delivery.Body)
}

func TestBatch_EncodedPublish_DecodeConsumer_Retry(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	codec := codec.Default
	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher(grmqx.PublisherLog(test.Logger(), true), grmqx.EncodeMessage(codec, 10))
	callCount := atomic.Int32{}

	publish := map[string]any{
		"field":  "value",
		"field2": fake.It[string](),
		"field3": fake.It[int](),
	}
	rawPublish, err := json.Marshal(publish)
	require.NoError(err)

	handler := grmqx.NewResultBatchHandler(
		test.Logger(),
		batch_handler.SyncHandlerAdapterFunc(func(batch batch_handler.BatchItems) {
			callCount.Add(int32(len(batch)))

			for _, item := range batch {
				require.Equal(codec.Type(), item.Delivery.Source().ContentEncoding)
				require.Equal(rawPublish, item.Delivery.Body)
			}
			batch.RetryAll(errors.Errorf("some error"))
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
	consumer := consumerCfg.DefaultConsumer(
		handler,
		grmqx.DecodeMessage(codec, test.Logger()),
		grmqx.ConsumerLog(test.Logger(), true),
	)
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.ConsumerConfig())),
	)
	cli.Upgrade(config)

	for range 10 {
		err := pub.Publish(t.Context(), &amqp091.Publishing{Body: rawPublish})
		require.NoError(err)
	}

	require.Eventually(func() bool {
		return callCount.Load() == 40
	}, 2*time.Second, 100*time.Millisecond)

	require.EqualValues(40, callCount.Load())
	require.Empty(cli.QueueLength("test"))
	require.EqualValues(10, cli.QueueLength("test.DLQ"))

	expectedPublish, err := codec.EncodeBytes(rawPublish)
	require.NoError(err)

	delivery := cli.DrainMessage("test.DLQ")
	require.Equal(codec.Type(), delivery.ContentEncoding)
	require.Equal(expectedPublish, delivery.Body)
}

func TestBatch_PlainPublish_DecodeConsumer_DecodeFailed(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	codec := codec.Default
	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher(grmqx.PublisherLog(test.Logger(), true))

	publish := map[string]any{
		"field":  "value",
		"field2": fake.It[string](),
		"field3": fake.It[int](),
	}
	rawPublish, err := json.Marshal(publish)
	require.NoError(err)

	handler := grmqx.NewResultBatchHandler(
		test.Logger(),
		batch_handler.SyncHandlerAdapterFunc(func(batch batch_handler.BatchItems) {
			require.Fail("received delivery after failed decode")
			batch.AckAll()
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
	consumer := consumerCfg.DefaultConsumer(
		handler,
		grmqx.DecodeMessage(codec, test.Logger()),
		grmqx.ConsumerLog(test.Logger(), true),
	)
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.ConsumerConfig())),
	)
	cli.Upgrade(config)

	for range 10 {
		err := pub.Publish(t.Context(), &amqp091.Publishing{
			ContentEncoding: codec.Type(),
			Body:            rawPublish,
		})
		require.NoError(err)
	}

	require.Eventually(func() bool {
		return cli.QueueLength("test.DLQ") == 10
	}, 2*time.Second, 100*time.Millisecond)

	require.Empty(cli.QueueLength("test"))
	require.EqualValues(10, cli.QueueLength("test.DLQ"))

	delivery := cli.DrainMessage("test.DLQ")
	require.Equal(codec.Type(), delivery.ContentEncoding)
	require.Equal(rawPublish, delivery.Body)
}

func TestBatch_PlainPublish_DecodeConsumer(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	codec := codec.Default
	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher(grmqx.PublisherLog(test.Logger(), true))
	callCount := atomic.Int32{}

	publish := map[string]any{
		"field":  "value",
		"field2": fake.It[string](),
		"field3": fake.It[int](),
	}
	rawPublish, err := json.Marshal(publish)
	require.NoError(err)

	handler := grmqx.NewResultBatchHandler(
		test.Logger(),
		batch_handler.SyncHandlerAdapterFunc(func(batch batch_handler.BatchItems) {
			callCount.Add(int32(len(batch)))

			for _, item := range batch {
				require.Empty(item.Delivery.Source().ContentEncoding)
				require.Equal(rawPublish, item.Delivery.Body)
			}
			batch.AckAll()
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
	consumer := consumerCfg.DefaultConsumer(
		handler,
		grmqx.DecodeMessage(codec, test.Logger()),
		grmqx.ConsumerLog(test.Logger(), true),
	)
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.ConsumerConfig())),
	)
	cli.Upgrade(config)

	for range 10 {
		err := pub.Publish(t.Context(), &amqp091.Publishing{Body: rawPublish})
		require.NoError(err)
	}

	require.Eventually(func() bool {
		return callCount.Load() == 10
	}, 2*time.Second, 100*time.Millisecond)

	require.EqualValues(10, callCount.Load())
	require.Empty(cli.QueueLength("test"))
	require.Empty(cli.QueueLength("test.DLQ"))
}

func TestBatch_PlainPublish_DecodeConsumer_Retry(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	codec := codec.Default
	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher(grmqx.PublisherLog(test.Logger(), true))
	callCount := atomic.Int32{}

	publish := map[string]any{
		"field":  "value",
		"field2": fake.It[string](),
		"field3": fake.It[int](),
	}
	rawPublish, err := json.Marshal(publish)
	require.NoError(err)

	handler := grmqx.NewResultBatchHandler(
		test.Logger(),
		batch_handler.SyncHandlerAdapterFunc(func(batch batch_handler.BatchItems) {
			callCount.Add(int32(len(batch)))

			for _, item := range batch {
				require.Empty(item.Delivery.Source().ContentEncoding)
				require.Equal(rawPublish, item.Delivery.Body)
			}
			batch.RetryAll(errors.Errorf("some error"))
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
	consumer := consumerCfg.DefaultConsumer(
		handler,
		grmqx.DecodeMessage(codec, test.Logger()),
		grmqx.ConsumerLog(test.Logger(), true),
	)
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg.ConsumerConfig())),
	)
	cli.Upgrade(config)

	for range 10 {
		err := pub.Publish(t.Context(), &amqp091.Publishing{Body: rawPublish})
		require.NoError(err)
	}

	require.Eventually(func() bool {
		return callCount.Load() == 40
	}, 2*time.Second, 100*time.Millisecond)

	require.EqualValues(40, callCount.Load())
	require.Empty(cli.QueueLength("test"))
	require.EqualValues(10, cli.QueueLength("test.DLQ"))

	delivery := cli.DrainMessage("test.DLQ")
	require.Empty(delivery.ContentEncoding)
	require.Equal(rawPublish, delivery.Body)
}

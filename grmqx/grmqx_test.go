package grmqx_test

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
	pub1 := pubCfg1.DefaultPublisher(grmqx.PublisherLog(test.Logger(), true))
	consumerCfg1 := grmqx.Consumer{
		Queue: "queue1",
	}
	pubCfg2 := grmqx.Publisher{
		RoutingKey: "queue2",
	}
	pub2 := pubCfg2.DefaultPublisher(grmqx.PublisherLog(test.Logger(), true))
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
	consumer1 := consumerCfg1.DefaultConsumer(handler1, grmqx.ConsumerLog(test.Logger(), true))

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
	consumer2 := consumerCfg2.DefaultConsumer(handler2, grmqx.ConsumerLog(test.Logger(), true))

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
	err := cli.Upgrade(t.Context(), cfg)
	require.NoError(err)

	ctx := requestid.ToContext(t.Context(), expectedRequestId)
	ctx = log.ToContext(ctx, log.String(requestid.LogKey, expectedRequestId))
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
		QueueArgs: map[string]any{
			"x-single-active-consumer": true,
		},
	}
	consumer := consumerCfg.DefaultConsumer(handler, grmqx.ConsumerLog(test.Logger(), true))
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg)),
	)
	cli.Upgrade(config)

	err := pub.Publish(t.Context(), &amqp091.Publishing{})
	require.NoError(err)

	time.Sleep(2 * time.Second)

	require.EqualValues(4, callCount.Load())
	require.EqualValues(1, cli.QueueLength("test.DLQ"))
}

func TestRecovery(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	pub := grmqx.Publisher{
		RoutingKey: "test",
	}.DefaultPublisher()

	handler := grmqx.NewResultHandler(
		test.Logger(),
		handler.SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) handler.Result {
			panic(errors.New("test panic error"))
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
	consumer := consumerCfg.DefaultConsumer(handler, grmqx.ConsumerLog(test.Logger(), true))
	cli := grmqt.New(test)
	config := grmqx.NewConfig("",
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(pub),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg)),
	)
	cli.Upgrade(config)

	err := pub.Publish(t.Context(), &amqp091.Publishing{})
	require.NoError(err)

	time.Sleep(2 * time.Second)

	require.EqualValues(1, cli.QueueLength("test.DLQ"))
}

func TestQueuesDelete(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	const queueName = "queue_to_delete"
	consumerCfg := grmqx.Consumer{
		Queue: queueName,
	}

	testCli := grmqt.New(test)
	config := grmqx.NewConfig(
		testCli.ConnectionConfig().Url(),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg)),
	)
	err := testCli.GrmqxCli.Upgrade(t.Context(), config)
	require.NoError(err)

	require.True(queueExists(t, testCli.ConnectionConfig().Url(), queueName))

	err = testCli.GrmqxCli.DeleteQueues(t.Context(), queueName)
	require.NoError(err)

	require.False(queueExists(t, testCli.ConnectionConfig().Url(), queueName))
}

func TestQueuesDeleteUninitializedClient(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	cli := grmqx.New(test.Logger())
	err := cli.DeleteQueues(t.Context(), "some-queue")
	require.Contains(err.Error(), "client is not initialized")
}

func queueExists(t *testing.T, url string, queue string) bool {
	t.Helper()
	conn, err := amqp091.Dial(url)
	if err != nil {
		t.Fatalf("dial rabbit mq: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("open channel: %v", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclarePassive(queue, false, false, false, false, nil)
	return err == nil
}

func TestQueueInspect(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	testCli := grmqt.New(test)
	config := grmqx.NewConfig(
		testCli.ConnectionConfig().Url(),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(grmqx.Consumer{Queue: "test_queue_inspect"})),
	)
	err := testCli.GrmqxCli.Upgrade(t.Context(), config)
	require.NoError(err)

	cli := grmqx.New(test.Logger())
	t.Cleanup(func() {
		cli.Close()
	})
	err = cli.Upgrade(t.Context(), config)
	require.NoError(err)

	queue, err := cli.QueueInspect("test_queue_inspect")
	require.NoError(err)
	require.EqualValues("test_queue_inspect", queue.Name)

	_, err = cli.QueueInspect("")
	require.Error(err)

	queue, err = cli.QueueInspect("non_existent_queue_xyz")
	require.Error(err)
	require.Empty(queue.Name)

	disconnectedCli := grmqx.New(test.Logger())
	t.Cleanup(func() {
		disconnectedCli.Close()
	})
	_, err = disconnectedCli.QueueInspect("some_queue")
	require.ErrorContains(err, "client is not initialized")
}

func TestQueuesDeleteWithInspect(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	const existingQueue = "existing_queue"
	const nonExistentQueue = "non_existent_queue_xyz"

	testCli := grmqt.New(test)
	consumerCfg := grmqx.Consumer{
		Queue: existingQueue,
	}
	config := grmqx.NewConfig(
		testCli.ConnectionConfig().Url(),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg)),
	)

	err := testCli.GrmqxCli.Upgrade(t.Context(), config)
	require.NoError(err)

	require.True(queueExists(t, testCli.ConnectionConfig().Url(), existingQueue))

	emptyResult, err := testCli.GrmqxCli.DeleteQueuesWithInspect(t.Context())
	require.NoError(err)
	require.Empty(emptyResult)

	results, err := testCli.GrmqxCli.DeleteQueuesWithInspect(t.Context(), "")
	require.NoError(err)
	require.Len(results, 1)
	require.Error(results[""])

	results, err = testCli.GrmqxCli.DeleteQueuesWithInspect(t.Context(), existingQueue, nonExistentQueue, "")
	require.NoError(err)
	require.NotNil(results)
	require.Len(results, 3)
	require.Error(results[nonExistentQueue])
	require.Contains(results[nonExistentQueue].Error(), grmqx.ErrNotExistQueue.Error())
	require.Error(results[""])
	require.NoError(results[existingQueue])
}

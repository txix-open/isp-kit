package stompx_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-stomp/stomp/v3"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"github.com/txix-open/isp-kit/stompx"
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/handler"
	"github.com/txix-open/isp-kit/stompx/publisher"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/stompt"
)

const (
	testRecoverQueue = "test_recover_queue"
)

func TestRequestIdChain(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)
	logger := test.Logger()

	expectedRequestId := requestid.Next()

	cli := stompt.New(test)

	pubCfg1 := cli.PublisherConfig("queue1")
	pub1 := stompx.DefaultPublisher(pubCfg1, stompx.PublisherLog(logger, true))
	consumerCfg1 := cli.ConsumerConfig("queue1")

	pubCfg2 := cli.PublisherConfig("queue2")
	pub2 := stompx.DefaultPublisher(pubCfg2, stompx.PublisherLog(logger, true))
	consumerCfg2 := cli.ConsumerConfig("queue2")

	handler1 := stompx.NewResultHandler(
		logger,
		handler.AdapterFunc(func(ctx context.Context, msg *stomp.Message) handler.Result {
			err := pub2.Publish(ctx, &publisher.Message{})
			require.NoError(err)
			return handler.Ack()
		}),
	)
	consumerConfig1 := stompx.DefaultConsumer(consumerCfg1, handler1, logger, stompx.ConsumerLog(logger, true))
	consumer1 := consumer.NewWatcher(consumerConfig1)

	await := make(chan struct{})
	handler2 := stompx.NewResultHandler(
		test.Logger(),
		handler.AdapterFunc(func(ctx context.Context, msg *stomp.Message) handler.Result {
			requestId := requestid.FromContext(ctx)
			require.EqualValues(expectedRequestId, requestId)
			close(await)
			return handler.Ack()
		}),
	)
	consumerConfig2 := stompx.DefaultConsumer(consumerCfg2, handler2, logger, stompx.ConsumerLog(logger, true))
	consumer2 := consumer.NewWatcher(consumerConfig2)

	stompCli := stompx.New(logger)
	t.Cleanup(func() {
		err := stompCli.Close()
		require.NoError(err)
	})

	config := stompx.NewConfig(
		stompx.WithConsumers(consumer1, consumer2),
	)

	stompCli.UpgradeAndServe(t.Context(), config)

	ctx := requestid.ToContext(t.Context(), expectedRequestId)
	ctx = log.ToContext(ctx, log.String(requestid.LogKey, expectedRequestId))
	err := pub1.Publish(ctx, &publisher.Message{})
	require.NoError(err)

	select {
	case <-await:
	case <-time.After(5 * time.Second):
		require.Fail("handler wasn't called")
	}
}

func TestRecover(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)
	logger := test.Logger()

	cli := stompt.New(test)

	pubCfg1 := cli.PublisherConfig(testRecoverQueue)
	pub1 := stompx.DefaultPublisher(pubCfg1, stompx.PublisherLog(logger, true))

	consumerCfg1 := cli.ConsumerConfig(testRecoverQueue)

	counter := 0
	handler1 := stompx.NewResultHandler(
		logger,
		handler.AdapterFunc(func(ctx context.Context, msg *stomp.Message) handler.Result {
			if counter == 0 {
				counter++
				panic(errors.New("test panic error"))
			}

			require.NotZero(counter)
			return handler.Ack()
		}),
	)
	consumerConfig1 := stompx.DefaultConsumer(consumerCfg1, handler1, logger, stompx.ConsumerLog(logger, true))
	consumer1 := consumer.NewWatcher(consumerConfig1)

	stompCli := stompx.New(logger)
	t.Cleanup(func() {
		err := stompCli.Close()
		require.NoError(err)
	})

	config := stompx.NewConfig(
		stompx.WithConsumers(consumer1),
	)
	stompCli.UpgradeAndServe(t.Context(), config)

	err := pub1.Publish(t.Context(), &publisher.Message{
		Body: []byte("test msg"),
	})
	require.NoError(err)

	time.Sleep(1 * time.Second)

	err = pub1.Publish(t.Context(), &publisher.Message{
		Body: []byte("test msg"),
	})
	require.NoError(err)
}

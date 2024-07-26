package tests

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax"
	kafkaHandler "github.com/txix-open/isp-kit/kafkax/handler"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"github.com/txix-open/isp-kit/test"
)

const (
	testRequestIdTopic = "test_topic"
	testRetryTopic     = "test_retry_topic"
)

func TestRequestIdChain(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)
	await := make(chan struct{})

	host := test.Config().Optional().String("KAFKA_ADDRESS", "localhost:9093")
	expectedRequestId := requestid.Next()

	_ = MakeMockConn(test, ConnectionConfig{
		Topic:   testRequestIdTopic,
		Brokers: []string{host},
	})

	time.Sleep(1 * time.Second)

	pubCfg1 := kafkax.PublisherConfig{
		Hosts:           []string{host},
		Topic:           testRequestIdTopic,
		MaxMsgSizeMb:    1,
		WriteTimeoutSec: 0,
	}
	pub1 := pubCfg1.DefaultPublisher(test.Logger(), kafkax.PublisherLog(test.Logger()))

	consumerCfg1 := kafkax.ConsumerConfig{
		Brokers: []string{host},
		Topic:   testRequestIdTopic,
		GroupId: "kkd",
	}

	handler1 := kafkax.NewResultHandler(
		test.Logger(),
		kafkaHandler.SyncHandlerAdapterFunc(func(ctx context.Context, msg *kafka.Message) kafkaHandler.Result {
			requestId := requestid.FromContext(ctx)
			require.EqualValues(expectedRequestId, requestId)
			require.EqualValues("test message", string(msg.Value))

			close(await)
			return kafkaHandler.Commit()
		}),
	)

	cons1 := consumerCfg1.DefaultConsumer(test.Logger(), handler1, kafkax.ConsumerLog(test.Logger()))

	kafkaBrokerConfig := kafkax.NewConfig(
		kafkax.WithPublishers(pub1),
		kafkax.WithConsumers(cons1),
	)

	client := kafkax.New(test.Logger())
	client.UpgradeAndServe(context.Background(), kafkaBrokerConfig)

	ctx := requestid.ToContext(context.Background(), expectedRequestId)
	ctx = log.ToContext(ctx, log.String("requestId", expectedRequestId))

	err := pub1.Publish(ctx, &kafka.Message{
		Value: []byte("test message"),
	})
	require.NoError(err)

	select {
	case <-await:
	case <-time.After(20 * time.Second):
		require.Fail("handler wasn't called")
	}
}

func TestRetry(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)
	await := make(chan struct{})

	host := test.Config().Optional().String("KAFKA_ADDRESS", "localhost:9093")
	counter := 0

	_ = MakeMockConn(test, ConnectionConfig{
		Topic:   testRetryTopic,
		Brokers: []string{host},
	})

	time.Sleep(1 * time.Second)

	pubCfg1 := kafkax.PublisherConfig{
		Hosts:           []string{host},
		Topic:           testRetryTopic,
		MaxMsgSizeMb:    1,
		WriteTimeoutSec: 0,
	}
	pub1 := pubCfg1.DefaultPublisher(test.Logger(), kafkax.PublisherLog(test.Logger()))

	consumerCfg1 := kafkax.ConsumerConfig{
		Brokers: []string{host},
		Topic:   testRetryTopic,
		GroupId: "kkd",
	}

	handler1 := kafkax.NewResultHandler(
		test.Logger(),
		kafkaHandler.SyncHandlerAdapterFunc(func(ctx context.Context, msg *kafka.Message) kafkaHandler.Result {
			if counter != 3 {
				counter++
				return kafkaHandler.Retry(1*time.Second, errors.New("some error"))
			}

			close(await)
			return kafkaHandler.Commit()
		}),
	)
	cons1 := consumerCfg1.DefaultConsumer(test.Logger(), handler1, kafkax.ConsumerLog(test.Logger()))

	kafkaBrokerConfig := kafkax.NewConfig(
		kafkax.WithPublishers(pub1),
		kafkax.WithConsumers(cons1),
	)

	client := kafkax.New(test.Logger())
	client.UpgradeAndServe(context.Background(), kafkaBrokerConfig)

	err := pub1.Publish(context.Background(), &kafka.Message{
		Value: []byte("test message"),
	})
	require.NoError(err)

	select {
	case <-await:
		require.EqualValues(3, counter)
	case <-time.After(20 * time.Second):
		require.Fail("handler wasn't called")
	}
}

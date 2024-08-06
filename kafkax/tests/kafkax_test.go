package tests

import (
	"context"
	"strconv"
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

	host := test.Config().Optional().String("KAFKA_ADDRESS", "10.2.4.244:9092")
	expectedRequestId := requestid.Next()

	_ = MakeMockConn(test, ConnectionConfig{
		Topic:    testRequestIdTopic,
		Brokers:  []string{host},
		Username: "kkd",
		Password: "iwrniL1FQbRRQuU3bWJVNluY",
	})

	time.Sleep(1 * time.Second)

	pubCfg1 := kafkax.PublisherConfig{
		Addresses:       []string{host},
		Topic:           testRequestIdTopic,
		MaxMsgSizeMb:    1,
		WriteTimeoutSec: 0,
		Auth: &kafkax.Auth{
			Username: "kkd",
			Password: "iwrniL1FQbRRQuU3bWJVNluY",
		},
	}
	pub1 := pubCfg1.DefaultPublisher(test.Logger(), kafkax.PublisherLog(test.Logger()))

	consumerCfg1 := kafkax.ConsumerConfig{
		Addresses: []string{host},
		Topic:     testRequestIdTopic,
		GroupId:   "kkd",
		Auth: &kafkax.Auth{
			Username: "kkd",
			Password: "iwrniL1FQbRRQuU3bWJVNluY",
		},
		Concurrency: 3,
	}

	handler1 := kafkax.NewResultHandler(
		test.Logger(),
		kafkaHandler.SyncHandlerAdapterFunc(func(ctx context.Context, msg *kafka.Message) kafkaHandler.Result {
			requestId := requestid.FromContext(ctx)
			require.EqualValues(expectedRequestId, requestId)
			require.EqualValues("test message", string(msg.Value))

			await <- struct{}{}
			return kafkaHandler.Commit()
		}),
	)

	cons1 := consumerCfg1.DefaultConsumer(test.Logger(), handler1, kafkax.ConsumerLog(test.Logger()))

	kafkaBrokerConfig := kafkax.NewConfig(
		kafkax.WithPublishers(pub1),
		kafkax.WithConsumers(cons1),
	)

	client := kafkax.New(test.Logger())
	err := client.UpgradeAndServe(context.Background(), kafkaBrokerConfig)
	require.NoError(err)

	time.Sleep(500 * time.Millisecond)

	ctx := requestid.ToContext(context.Background(), expectedRequestId)
	ctx = log.ToContext(ctx, log.String("requestId", expectedRequestId))

	for i := 1; i <= 5; i++ {
		err = pub1.Publish(ctx, kafka.Message{
			Key:   []byte(strconv.Itoa(i)),
			Value: []byte("test message"),
		})
		require.NoError(err)
	}

	counter := 0

	for {
		select {
		case <-await:
			counter++
			if counter == 5 {
				client.Close()
				return
			}
		case <-time.After(20 * time.Second):
			require.Fail("handler wasn't called")
		}
	}
}

func TestRetry(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)
	await := make(chan struct{})

	host := test.Config().Optional().String("KAFKA_ADDRESS", "10.2.4.244:9092")
	counter := 0

	_ = MakeMockConn(test, ConnectionConfig{
		Topic:    testRetryTopic,
		Brokers:  []string{host},
		Username: "kkd",
		Password: "iwrniL1FQbRRQuU3bWJVNluY",
	})

	time.Sleep(1 * time.Second)

	pubCfg1 := kafkax.PublisherConfig{
		Addresses:       []string{host},
		Topic:           testRetryTopic,
		MaxMsgSizeMb:    1,
		WriteTimeoutSec: 0,
		Auth: &kafkax.Auth{
			Username: "kkd",
			Password: "iwrniL1FQbRRQuU3bWJVNluY",
		},
	}
	pub1 := pubCfg1.DefaultPublisher(test.Logger(), kafkax.PublisherLog(test.Logger()))

	consumerCfg1 := kafkax.ConsumerConfig{
		Addresses: []string{host},
		Topic:     testRetryTopic,
		GroupId:   "kkd",
		Auth: &kafkax.Auth{
			Username: "kkd",
			Password: "iwrniL1FQbRRQuU3bWJVNluY",
		},
	}

	handler1 := kafkax.NewResultHandler(
		test.Logger(),
		kafkaHandler.SyncHandlerAdapterFunc(func(ctx context.Context, msg *kafka.Message) kafkaHandler.Result {
			if counter != 3 {
				counter++
				return kafkaHandler.Retry(1*time.Second, errors.New("some error"))
			}

			defer close(await)
			return kafkaHandler.Commit()
		}),
	)
	cons1 := consumerCfg1.DefaultConsumer(test.Logger(), handler1, kafkax.ConsumerLog(test.Logger()))

	kafkaBrokerConfig := kafkax.NewConfig(
		kafkax.WithPublishers(pub1),
		kafkax.WithConsumers(cons1),
	)

	client := kafkax.New(test.Logger())
	err := client.UpgradeAndServe(context.Background(), kafkaBrokerConfig)
	require.NoError(err)

	err = pub1.Publish(context.Background(), kafka.Message{
		Value: []byte("test message"),
	})
	require.NoError(err)

	select {
	case <-await:
		require.EqualValues(3, counter)
		client.Close()
	case <-time.After(25 * time.Second):
		require.Fail("handler wasn't called")
	}
}

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax"
	kafkaHandler "github.com/txix-open/isp-kit/kafkax/handler"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"github.com/txix-open/isp-kit/test"
)

const testTopic = "test_topic"

func TestRequestIdChain(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)
	await := make(chan struct{})

	host := test.Config().Optional().String("KAFKA_ADDRESS", "10.2.4.244:9093")
	expectedRequestId := requestid.Next()

	pubCfg1 := kafkax.PublisherConfig{
		Hosts:           []string{host},
		Topic:           testTopic,
		MaxMsgSizeMb:    1,
		WriteTimeoutSec: 0,
		Auth: &kafkax.Auth{
			Username: "kkd",
			Password: "iwrniL1FQbRRQuU3bWJVNluY",
		},
	}
	pub1 := pubCfg1.DefaultPublisher(test.Logger(), kafkax.PublisherLog(test.Logger()))

	consumerCfg1 := kafkax.ConsumerConfig{
		Brokers: []string{host},
		Topic:   testTopic,
		Auth: &kafkax.Auth{
			Username: "kkd",
			Password: "iwrniL1FQbRRQuU3bWJVNluY",
		},
	}

	handler1 := kafkax.NewResultHandler(
		test.Logger(),
		kafkaHandler.SyncHandlerAdapterFunc(func(ctx context.Context, msg *kafka.Message) kafkaHandler.Result {
			requestId := requestid.FromContext(ctx)
			require.EqualValues(expectedRequestId, requestId)

			close(await)

			return kafkaHandler.Commit()
		}),
	)

	cons1 := consumerCfg1.DefaultConsumer(test.Logger(), handler1, kafkax.ConsumerLog(test.Logger()))

	conn := MakeMockConn(test,
		ConnectionConfig{
			Brokers:  []string{host},
			Topic:    testTopic,
			Username: "kkd",
			Password: "iwrniL1FQbRRQuU3bWJVNluY",
		})

	brokers, err := conn.Brokers()
	require.NoError(err)

	fmt.Println("brokers", brokers)

	kafkaBrokerConfig := kafkax.NewConfig(
		kafkax.WithPublishers(pub1),
		kafkax.WithConsumers(cons1),
	)

	client := kafkax.New(test.Logger())
	client.UpgradeAndServe(context.Background(), kafkaBrokerConfig)

	ctx := requestid.ToContext(context.Background(), expectedRequestId)
	ctx = log.ToContext(ctx, log.String("requestId", expectedRequestId))

	err = pub1.Publish(ctx, &kafka.Message{Value: []byte("publisher with middlewares")})
	require.NoError(err)

	select {
	case <-await:
	case <-time.After(5 * time.Second):
		require.Fail("handler wasn't called")
	}
}

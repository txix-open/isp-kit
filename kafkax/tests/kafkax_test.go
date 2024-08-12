package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	kafkaHandler "github.com/txix-open/isp-kit/kafkax/handler"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"github.com/txix-open/isp-kit/test"
)

const (
	testRequestIdTopic = "test_topic"
	groupIdRequestId   = "testRequestId"
	testRetryTopic     = "test_retry_topic"
	groupIdRetry       = "testRetry"
)

func TestRequestIdChain(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)
	await := make(chan struct{})

	testKafka := NewKafka(test, testRequestIdTopic)

	time.Sleep(500 * time.Millisecond)

	pubCfg := testKafka.PublisherConfig(testRequestIdTopic)
	pub1 := pubCfg.DefaultPublisher(test.Logger(), kafkax.PublisherLog(test.Logger()))

	consumerCfg := testKafka.ConsumerConfig(testRequestIdTopic, groupIdRequestId)
	consumerCfg.Concurrency = 3

	expectedRequestId := requestid.Next()
	handler1 := kafkax.NewResultHandler(
		test.Logger(),
		kafkaHandler.SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) kafkaHandler.Result {
			requestId := requestid.FromContext(ctx)
			require.EqualValues(expectedRequestId, requestId)
			require.EqualValues("test message", string(delivery.Source().Value))

			await <- struct{}{}
			return kafkaHandler.Commit()
		}),
	)

	cons1 := consumerCfg.DefaultConsumer(test.Logger(), handler1, kafkax.ConsumerLog(test.Logger()))

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

	testKafka := NewKafka(test, testRetryTopic)

	time.Sleep(500 * time.Millisecond)

	pubCfg := testKafka.PublisherConfig(testRetryTopic)
	pub1 := pubCfg.DefaultPublisher(test.Logger(), kafkax.PublisherLog(test.Logger()))

	consumerCfg := testKafka.ConsumerConfig(testRetryTopic, groupIdRetry)

	counter := 0
	handler1 := kafkax.NewResultHandler(
		test.Logger(),
		kafkaHandler.SyncHandlerAdapterFunc(func(ctx context.Context, delivery *consumer.Delivery) kafkaHandler.Result {
			if counter != 3 {
				counter++
				return kafkaHandler.Retry(1*time.Second, errors.New("some error"))
			}

			defer close(await)
			return kafkaHandler.Commit()
		}),
	)
	cons1 := consumerCfg.DefaultConsumer(test.Logger(), handler1, kafkax.ConsumerLog(test.Logger()))

	kafkaBrokerConfig := kafkax.NewConfig(
		kafkax.WithPublishers(pub1),
		kafkax.WithConsumers(cons1),
	)

	client := kafkax.New(test.Logger())
	err := client.UpgradeAndServe(context.Background(), kafkaBrokerConfig)
	require.NoError(err)

	time.Sleep(500 * time.Millisecond)

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

func TestReadWrite(t *testing.T) {
	t.Parallel()
	test, require := test.New(t)

	testKafka := NewKafka(test, "test_read_write")

	time.Sleep(500 * time.Millisecond)

	testKafka.WriteMessages(kafka.Message{
		Topic: testKafka.Topic(),
		Value: []byte("test message"),
	})

	time.Sleep(200 * time.Millisecond)

	msg := testKafka.ReadMessage()

	require.EqualValues([]byte("test message"), msg.Value)
	require.EqualValues(testKafka.Topic(), msg.Topic)

	testKafka.CommitMessages(msg)

	testKafka.WriteMessages(kafka.Message{
		Topic: testKafka.Topic(),
		Value: []byte("test message 2"),
	})

	time.Sleep(200 * time.Millisecond)

	msg = testKafka.ReadMessage()

	require.EqualValues([]byte("test message 2"), msg.Value)
	require.EqualValues(testKafka.Topic(), msg.Topic)
}

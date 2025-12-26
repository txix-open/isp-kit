package kafkat

import (
	"context"
	"fmt"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"os"
	"time"

	"github.com/txix-open/isp-kit/kafkax"
	"github.com/txix-open/isp-kit/test"
)

type Kafka struct {
	test     *test.Test
	client   *kgo.Client
	adm      *kadm.Client
	address  string
	username string
	password string
	topics   []string
}

func NewKafka(t *test.Test) *Kafka {
	addr := t.Config().Optional().String("KAFKA_ADDRESS", "127.0.0.1:9092")
	username := t.Config().Optional().String("KAFKA_USERNAME", "user")
	password := t.Config().Optional().String("KAFKA_PASSWORD", "password")

	c, err := kgo.NewClient(
		kgo.SeedBrokers(addr),
		kgo.ProduceRequestTimeout(500*time.Millisecond), //nolint:mnd
		kgo.MaxBufferedRecords(1),
		kgo.SASL(plain.Auth{
			User: username,
			Pass: password,
		}.AsMechanism()),
		kgo.WithLogger(kgo.BasicLogger(os.Stderr, kgo.LogLevelError, func() string {
			return "kafka client"
		})))
	t.Assert().NoError(err)

	adm := kadm.NewClient(c)

	cli := &Kafka{
		test:     t,
		client:   c,
		adm:      adm,
		address:  addr,
		username: username,
		password: password,
		topics:   make([]string, 0),
	}

	t.T().Cleanup(func() {
		cli.deleteTopics()
		cli.client.Close()
	})

	return cli
}

// WriteMessages публикует сообщения в топик, указанный в сообщении
func (k *Kafka) WriteMessages(msgs ...*kgo.Record) {
	results := k.client.ProduceSync(context.Background(), msgs...)
	for _, result := range results {
		if result.Err != nil {
			k.test.Assert().NoError(fmt.Errorf("failed to produce message: %w", result.Err))
		}
	}
}

func (k *Kafka) ReadMessage(topic string, offset int64) *kgo.Record {
	saslMechanism := plain.Auth{
		User: k.username,
		Pass: k.password,
	}.AsMechanism()

	opts := []kgo.Opt{
		kgo.SeedBrokers(k.address),
		kgo.SASL(saslMechanism),
		kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
			topic: {0: kgo.NewOffset().At(offset)},
		}),
	}

	cl, err := kgo.NewClient(opts...)
	k.test.Assert().NoError(err)
	defer cl.Close()

	ctx := context.Background()
	fetches := cl.PollFetches(ctx)

	k.test.Assert().NoError(fetches.Err())

	iter := fetches.RecordIter()
	if !iter.Done() {
		return iter.Next()
	}

	return nil
}

func (k *Kafka) Address() string {
	return k.address
}

func (k *Kafka) CreateDefaultTopic(topic string) {
	_, err := k.adm.CreateTopic(context.Background(), 1, -1, nil, topic)
	k.test.Assert().NoError(err)
	k.topics = append(k.topics, topic)
}

func (k *Kafka) PublisherConfig(topic string) kafkax.PublisherConfig {
	return kafkax.PublisherConfig{
		Addresses:             []string{k.address},
		Topic:                 topic,
		BatchSizePerPartition: 1,
		Auth: &kafkax.Auth{
			Username: k.username,
			Password: k.password,
		},
	}
}

func (k *Kafka) ConsumerConfig(topic, groupId string) kafkax.ConsumerConfig {
	return kafkax.ConsumerConfig{
		Addresses:   []string{k.address},
		Topic:       topic,
		GroupId:     groupId,
		Concurrency: 1,
		Auth: &kafkax.Auth{
			Username: k.username,
			Password: k.password,
		},
	}
}

func (k *Kafka) deleteTopics() {
	_, err := k.adm.DeleteTopics(context.Background(), k.topics...)
	k.test.Assert().NoError(err)
}

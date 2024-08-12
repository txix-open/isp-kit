// nolint:gomnd
package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/txix-open/isp-kit/kafkax"
	"github.com/txix-open/isp-kit/test"
)

type Kafka struct {
	test     *test.Test
	manager  *kafka.Conn
	writer   *kafka.Writer
	reader   *kafka.Reader
	address  string
	username string
	password string
	topics   []string
}

func NewKafka(t *test.Test, topic string) *Kafka {
	addr := t.Config().Optional().String("KAFKA_ADDRESS", "127.0.0.1:9092")
	username := t.Config().Optional().String("KAFKA_USERNAME", "user")
	password := t.Config().Optional().String("KAFKA_PASSWORD", "password")

	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second, //nolint:mnd
		DualStack: true,
		SASLMechanism: plain.Mechanism{
			Username: username,
			Password: password,
		},
	}

	conn, err := dialer.Dial("tcp", addr)
	t.Assert().NoError(err)

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: -1,
	})
	t.Assert().NoError(err)

	w := &kafka.Writer{
		Addr:         kafka.TCP(addr),
		BatchTimeout: 500 * time.Millisecond, //nolint:mnd
		BatchSize:    2,
		Transport: &kafka.Transport{
			SASL: kafkax.PlainAuth(&kafkax.Auth{
				Username: username,
				Password: password,
			}),
		},
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			t.Logger().Error(context.Background(), "kafka publisher: "+fmt.Sprintf(s, i...))
		}),
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{addr},
		GroupID: "test",
		Topic:   topic,
		Dialer: &kafka.Dialer{
			DualStack: true,
			Timeout:   10 * time.Second,
			SASLMechanism: kafkax.PlainAuth(&kafkax.Auth{
				Username: username,
				Password: password,
			}),
		},
		MinBytes: 1,
		MaxBytes: 64 * 1024 * 1024,
		ErrorLogger: kafka.LoggerFunc(func(s string, i ...interface{}) {
			t.Logger().Error(context.Background(), "kafka consumer: "+fmt.Sprintf(s, i...))
		}),
	})

	topics := []string{topic}

	t.T().Cleanup(func() {
		err = r.Close()
		t.Assert().NoError(err)

		err = w.Close()
		t.Assert().NoError(err)

		err = conn.DeleteTopics(topics...)
		t.Assert().NoError(err)

		err = conn.Close()
		t.Assert().NoError(err)
	})

	return &Kafka{
		test:     t,
		manager:  conn,
		writer:   w,
		reader:   r,
		address:  addr,
		username: username,
		password: password,
		topics:   topics,
	}
}

// WriteMessages публикует сообщения в топик, указанный в сообщении
func (k *Kafka) WriteMessages(msgs ...kafka.Message) {
	err := k.writer.WriteMessages(context.Background(), msgs...)
	k.test.Assert().NoError(err)
}

// ReadMessage читает сообщения из топика, переданного в NewKafka()
func (k *Kafka) ReadMessage() kafka.Message {
	msg, err := k.reader.ReadMessage(context.Background())
	k.test.Assert().NoError(err)

	return msg
}

func (k *Kafka) CommitMessages(msgs ...kafka.Message) {
	err := k.reader.CommitMessages(context.Background(), msgs...)
	k.test.Assert().NoError(err)
}

func (k *Kafka) Address() string {
	return k.address
}

// Topic возвращает название топика, указанное при вызове NewKafka()
func (k *Kafka) Topic() string {
	return k.topics[0]
}

func (k *Kafka) CreateDefaultTopic(topic string) {
	err := k.manager.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: -1,
	})
	k.test.Assert().NoError(err)
	k.topics = append(k.topics, topic)
}

func (k *Kafka) PublisherConfig(topic string) kafkax.PublisherConfig {
	return kafkax.PublisherConfig{
		Addresses: []string{k.address},
		Topic:     topic,
		BatchSize: 1,
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

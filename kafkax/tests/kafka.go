// nolint:gomnd
package tests

import (
	"context"
	"fmt"
	"io"
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
		BatchSize:    1,
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

	t.T().Cleanup(func() {
		err = w.Close()
		t.Assert().NoError(err)

		err = conn.DeleteTopics(topic)
		t.Assert().NoError(err)

		err = conn.Close()
		t.Assert().NoError(err)
	})

	return &Kafka{
		test:     t,
		manager:  conn,
		writer:   w,
		address:  addr,
		username: username,
		password: password,
		topics:   []string{topic},
	}
}

// WriteMessages публикует сообщения в топик, указанный в сообщении
func (k *Kafka) WriteMessages(msgs ...kafka.Message) {
	err := k.writer.WriteMessages(context.Background(), msgs...)
	k.test.Assert().NoError(err)
}

func (k *Kafka) ReadMessage(topic string, offset int64) kafka.Message {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second, //nolint:mnd
		DualStack: true,
		SASLMechanism: kafkax.PlainAuth(&kafkax.Auth{
			Username: k.username,
			Password: k.password,
		}),
	}

	l, err := dialer.DialLeader(context.Background(), "tcp", k.address, topic, 0)
	k.test.Assert().NoError(err)

	_, err = l.Seek(offset, io.SeekStart)
	k.test.Assert().NoError(err)

	msg, err := l.ReadMessage(64 * 1024 * 1024)
	k.test.Assert().NoError(err)

	err = l.Close()
	k.test.Assert().NoError(err)

	return msg
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

// DeleteTopics для удаления вручную созданных топиков
func (k *Kafka) DeleteTopics() {
	if len(k.topics[1:]) != 0 {
		err := k.manager.DeleteTopics(k.topics[1:]...)
		k.test.Assert().NoError(err)
	}
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

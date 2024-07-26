// nolint:gomnd
package tests

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/txix-open/isp-kit/kafkax"

	"github.com/txix-open/isp-kit/test"
)

type Kafka struct {
	manager *kafka.Conn
	writer  *kafka.Writer
	brokers []string
	test    *test.Test
}

type ConnectionConfig struct {
	Brokers  []string
	Topic    string
	Username string
	Password string
}

func MakeMockConn(t *test.Test, cfg ConnectionConfig) *kafka.Conn {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second, //nolint:mnd
		DualStack: true,
		SASLMechanism: plain.Mechanism{
			Username: cfg.Username,
			Password: cfg.Password,
		},
	}

	conn, err := dialer.Dial("tcp", cfg.Brokers[0])
	t.Assert().NoError(err)

	err = conn.CreateTopics(kafka.TopicConfig{
		Topic:             cfg.Topic,
		NumPartitions:     1,
		ReplicationFactor: -1,
	})

	//conn, err := dialer.DialLeader(context.Background(), "tcp", cfg.Brokers[0], topicName, 0)
	t.Assert().NoError(err)

	t.T().Cleanup(func() {
		err := conn.DeleteTopics(cfg.Topic)
		t.Assert().NoError(err)
	})

	return conn
}

func CreateTestTopic(t *test.Test, conn *kafka.Conn, topic string) {
	err := conn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: -1,
	})
	t.Assert().NoError(err)

	t.T().Cleanup(func() {
		err := conn.DeleteTopics(topic)
		t.Assert().NoError(err)
	})
}

func NewKafka(t *test.Test, auth *kafkax.Auth) *Kafka {
	addr := t.Config().Optional().String("KAFKA_ADDRESS", "10.2.4.244:9093")
	c, err := kafka.Dial("tcp", addr)
	t.Assert().NoError(err)
	t.T().Cleanup(func() {
		err := c.Close()
		t.Assert().NoError(err)
	})

	w := &kafka.Writer{
		Addr:         kafka.TCP(addr),
		BatchTimeout: 100 * time.Millisecond, //nolint:mnd
	}

	if auth != nil {
		w.Transport = &kafka.Transport{
			SASL: kafkax.PlainAuth(auth),
		}
	}

	t.T().Cleanup(func() {
		err := w.Close()
		t.Assert().NoError(err)
	})
	return &Kafka{
		writer:  w,
		manager: c,
		brokers: []string{addr},
		test:    t,
	}
}

func (k *Kafka) CreateTopics(topics ...kafka.TopicConfig) []string {
	toCreate := make([]kafka.TopicConfig, 0)
	toDelete := make([]string, 0)
	fullNames := make([]string, 0)
	totalPartitions := 0
	for _, topic := range topics {
		fullName := fmt.Sprintf("%s_%s", k.test.Id(), topic.Topic)
		toCreate = append(toCreate, kafka.TopicConfig{
			Topic:              fullName,
			NumPartitions:      topic.NumPartitions,
			ReplicationFactor:  topic.ReplicationFactor,
			ReplicaAssignments: topic.ReplicaAssignments,
			ConfigEntries:      topic.ConfigEntries,
		})
		if topic.NumPartitions == -1 {
			totalPartitions++
		} else {
			totalPartitions += topic.NumPartitions
		}
		toDelete = append(toDelete, fullName)
		fullNames = append(fullNames, fullName)
	}

	suffix := make([]byte, 4) //nolint:mnd
	_, err := rand.Read(suffix)
	k.test.Assert().NoError(err)
	readyProbeTopic := fmt.Sprintf("%s_%x", k.test.Id(), suffix)
	toCreate = append(toCreate, kafka.TopicConfig{
		Topic:             readyProbeTopic,
		NumPartitions:     1,
		ReplicationFactor: -1,
	})
	toDelete = append(toDelete, readyProbeTopic)

	err = k.manager.CreateTopics(toCreate...)
	k.test.Assert().NoError(err)

	k.test.T().Cleanup(func() {
		err := k.manager.DeleteTopics(toDelete...)
		k.test.Assert().NoError(err)
	})

	created := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd
	defer cancel()
	go func() {
		for {
			err := k.writer.WriteMessages(ctx, kafka.Message{
				Topic: readyProbeTopic,
				Key:   []byte("probe"),
			})
			if err == nil {
				close(created)
				return
			}
			select {
			case <-ctx.Done():
			case <-time.After(1 * time.Second):
			}
		}
	}()
	select {
	case <-created:
		return fullNames
	case <-ctx.Done():
		k.test.Assert().NoError(ctx.Err())
		return fullNames
	}
}

func (k *Kafka) WriteMessages(msgs ...kafka.Message) {
	err := k.writer.WriteMessages(context.Background(), msgs...)
	k.test.Assert().NoError(err)
}

func (k *Kafka) Brokers() []string {
	return k.brokers
}

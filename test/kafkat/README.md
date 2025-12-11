# Package `kafkat`

Пакет `kafkat` предназначен для упрощения написания тестов, взаимодействующих с Kafka, с использованием базового клиента-писателя и средств для управления топиками.

## Types

### Kafka

Структура `Kafka` предоставляет вспомогательные методы для создания и удаления топиков, публикации и чтения сообщений, а также генерации конфигураций Kafka-публикатора и консьюмера.

**Methods:**

#### `NewKafka(t *test.Test) *Kafka`

Создаёт экземпляр `Kafka`, инициализирует соединение с Kafka и писатель, а также регистрирует автоматическое удаление созданных топиков и закрытие соединений по завершению теста.

#### `(k *Kafka) WriteMessages(msgs ...*kgo.Record)`

Публикует переданные сообщения в соответствующие топики.

#### `(k *Kafka) ReadMessage(topic string, offset int64) kafka.Message`

Считывает одно сообщение из указанного топика с заданного смещения.

#### `(k *Kafka) Address() string`

Возвращает адрес Kafka-сервера.

#### `(k *Kafka) CreateDefaultTopic(topic string)`

Создаёт Kafka-топик с одной партицией и фактором репликации `-1`.

#### `(k *Kafka) PublisherConfig(topic string) kafkax.PublisherConfig`

Возвращает готовую конфигурацию Kafka-публикатора для заданного топика.

#### `(k *Kafka) ConsumerConfig(topic, groupId string) kafkax.ConsumerConfig`

Возвращает готовую конфигурацию Kafka-консьюмера для заданного топика и группы.

## Usage

### Example usage in test

```go
package mypkg_test

import (
	"testing"
	"github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkat"
	"github.com/txix-open/isp-kit/test"
)

func TestKafkaExample(t *testing.T) {
	testCtx := test.New(t)
	kafka := kafkat.NewKafka(testCtx)

	topic := "example-topic"
	kafka.CreateDefaultTopic(topic)

	msg := kafka.Message{Topic: topic, Value: []byte("test-message")}
	kafka.WriteMessages(msg)

	read := kafka.ReadMessage(topic, 0)
	testCtx.Assert().Equal("test-message", string(read.Value))
}
```

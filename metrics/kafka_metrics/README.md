# Package `kafka_metrics`

Пакет `kafka_metrics` предоставляет инструменты для сбора и регистрации метрик Kafka-потребителей и Kafka-публикаторов с использованием Prometheus.

## Types

### `ConsumerStorage`

Хранилище метрик Kafka-потребителя.

#### `func NewConsumerStorage(reg *metrics.Registry) *ConsumerStorage`

Создаёт новое хранилище метрик для Kafka-потребителя.

**Metrics:**

#### `kafka_consume_duration_ms`

Продолжительность обработки одного сообщения из топика

#### `kafka_consume_body_size`

Размер тела потреблённого сообщения

#### `kafka_consume_commit_count`

Счётчик коммитов сообщений

#### `kafka_consume_retry_count`

Счётчик ретраев сообщений

**Methods:**

#### `func (c *ConsumerStorage) ObserveConsumeDuration(consumerGroup, topic string, t time.Duration)`

Регистрирует задержку обработки одного сообщения.

#### `func (c *ConsumerStorage) ObserveConsumeMsgSize(consumerGroup, topic string, size int)`

Регистрирует размер тела потреблённого сообщения.

#### `func (c *ConsumerStorage) IncCommitCount(consumerGroup, topic string)`

Увеличивает счётчик коммитов сообщений.

#### `func (c *ConsumerStorage) IncRetryCount(consumerGroup, topic string)`

Увеличивает счётчик ретраев сообщений.

### `PublisherStorage`

Хранилище метрик Kafka-публикатора.

**Metrics:**

#### `kafka_publish_duration_ms`

Продолжительность публикации сообщения в топик.

#### `kafka_publish_body_size`

Размер тела публикуемого сообщения.

#### `kafka_publish_error_count`

Счётчик ошибок публикации.

**Methods:**

#### `func NewPublisherStorage(reg *metrics.Registry) *PublisherStorage`

Создаёт новое хранилище метрик для Kafka-публикатора.

#### `func (c *PublisherStorage) ObservePublishDuration(topic string, t time.Duration)`

Регистрирует задержку публикации сообщения.

#### `func (c *PublisherStorage) ObservePublishMsgSize(topic string, size int)`

Регистрирует размер тела публикуемого сообщения.

#### `func (c *PublisherStorage) IncPublishError(topic string)`

Увеличивает счётчик ошибок публикации.

## Prometheus metrics example

```
# HELP kafka_consume_duration_ms The latency of handling single message from topic
# TYPE kafka_consume_duration_ms summary
kafka_consume_duration_ms{consumerGroup="users-consumer",topic="user.events"} 5.1

# HELP kafka_consume_body_size The size of message body from queue
# TYPE kafka_consume_body_size summary
kafka_consume_body_size{consumerGroup="users-consumer",topic="user.events"} 1024

# HELP kafka_consume_commit_count Count of committed messages
# TYPE kafka_consume_commit_count counter
kafka_consume_commit_count{consumerGroup="users-consumer",topic="user.events"} 213

# HELP kafka_consume_retry_count Count of retried messages
# TYPE kafka_consume_retry_count counter
kafka_consume_retry_count{consumerGroup="users-consumer",topic="user.events"} 12

# HELP kafka_publish_duration_ms The latency of publishing messages to topic
# TYPE kafka_publish_duration_ms summary
kafka_publish_duration_ms{topic="user.events"} 3.4

# HELP kafka_publish_body_size The size of published message body to topic
# TYPE kafka_publish_body_size summary
kafka_publish_body_size{topic="user.events"} 512

# HELP kafka_publish_error_count Count error on publishing
# TYPE kafka_publish_error_count counter
kafka_publish_error_count{topic="user.events"} 1
```

## Usage

### Consumer

```go
consumerMetrics := kafka_metrics.NewConsumerStorage(metrics.DefaultRegistry)
...
consumerMetrics.ObserveConsumeDuration("users-consumer", "user.events", duration)
consumerMetrics.ObserveConsumeMsgSize("users-consumer", "user.events", size)
consumerMetrics.IncCommitCount("users-consumer", "user.events")
consumerMetrics.IncRetryCount("users-consumer", "user.events")
```

### Publisher

```go
publisherMetrics := kafka_metrics.NewPublisherStorage(metrics.DefaultRegistry)
...
publisherMetrics.ObservePublishDuration("user.events", duration)
publisherMetrics.ObservePublishMsgSize("user.events", size)
publisherMetrics.IncPublishError("user.events")
```

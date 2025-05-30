# Package `rabbitmq_metrics`

Пакет `rabbitmq_metrics` предоставляет метрики для мониторинга операций потребления и публикации сообщений в RabbitMQ с использованием Prometheus.

## Types

### ConsumerStorage

Хранилище метрик потребления сообщений.

**Metrics:**

#### `consume_duration_ms`

Продолжительность обработки сообщения.

#### `consume_body_size`

Размер тела сообщения.

#### `consume_requeue_count`

Количество повторно поставленных сообщений.

#### `consume_dlq_count`

Количество сообщений, отправленных в DLQ.

#### `consume_success_count`

Количество успешно обработанных сообщений.

#### `consume_retry_count`

Количество повторных попыток обработки.

**Methods:**

#### `func NewConsumerStorage(reg *metrics.Registry) *ConsumerStorage`

Создаёт новое хранилище метрик потребителя RabbitMQ.

#### `func (c *ConsumerStorage) ObserveConsumeDuration(exchange string, routingKey string, duration time.Duration)`

Регистрирует задержку обработки сообщения.

#### `func (c *ConsumerStorage) ObserveConsumeMsgSize(exchange string, routingKey string, size int)`

Регистрирует размер тела сообщения.

#### `func (c *ConsumerStorage) IncRequeueCount(exchange string, routingKey string)`

Увеличивает счётчик повторно поставленных сообщений.

#### `func (c *ConsumerStorage) IncDlqCount(exchange string, routingKey string)`

Увеличивает счётчик сообщений, отправленных в DLQ.

#### `func (c *ConsumerStorage) IncSuccessCount(exchange string, routingKey string)`

Увеличивает счётчик успешно обработанных сообщений.

#### `func (c *ConsumerStorage) IncRetryCount(exchange string, routingKey string)`

Увеличивает счётчик повторных попыток обработки.

### PublisherStorage

Хранилище метрик публикации сообщений.

**Metrics:**

#### `publish_duration_ms`

Продолжительность публикации сообщения.

#### `publish_body_size`

Размер тела публикуемого сообщения.

#### `publish_error_count`

Количество ошибок при публикации.

**Methods:**

#### `func NewPublisherStorage(reg *metrics.Registry) *PublisherStorage`

Создаёт новое хранилище метрик для публикации сообщений в RabbitMQ.

#### `func (c *PublisherStorage) ObservePublishDuration(exchange string, routingKey string, duration time.Duration)`

Регистрирует задержку публикации сообщения.

#### `func (c *PublisherStorage) ObservePublishMsgSize(exchange string, routingKey string, size int)`

Регистрирует размер публикуемого сообщения.

#### `func (c *PublisherStorage) IncPublishError(exchange string, routingKey string)`

Увеличивает счётчик ошибок публикации.

## Prometheus metrics example

```
# HELP rabbitmq_consume_duration_ms The latency of handling single message from queue
# TYPE rabbitmq_consume_duration_ms summary
rabbitmq_consume_duration_ms{exchange="exchange",routing_key="key"} 85.3

# HELP rabbitmq_consume_body_size The size of message body from queue
# TYPE rabbitmq_consume_body_size summary
rabbitmq_consume_body_size{exchange="exchange",routing_key="key"} 128

# HELP rabbitmq_consume_requeue_count Count of requeued messages
# TYPE rabbitmq_consume_requeue_count counter
rabbitmq_consume_requeue_count{exchange="exchange",routing_key="key"} 4

# HELP rabbitmq_consume_dlq_count Count of messages moved to DLQ
# TYPE rabbitmq_consume_dlq_count counter
rabbitmq_consume_dlq_count{exchange="exchange",routing_key="key"} 2

# HELP rabbitmq_consume_success_count Count of successful messages
# TYPE rabbitmq_consume_success_count counter
rabbitmq_consume_success_count{exchange="exchange",routing_key="key"} 97

# HELP rabbitmq_consume_retry_count Count of retried messages
# TYPE rabbitmq_consume_retry_count counter
rabbitmq_consume_retry_count{exchange="exchange",routing_key="key"} 7

# HELP rabbitmq_publish_duration_ms The latency of publishing single message to queue
# TYPE rabbitmq_publish_duration_ms summary
rabbitmq_publish_duration_ms{exchange="exchange",routing_key="key"} 19.7

# HELP rabbitmq_publish_body_size The size of published message body to queue
# TYPE rabbitmq_publish_body_size summary
rabbitmq_publish_body_size{exchange="exchange",routing_key="key"} 256

# HELP rabbitmq_publish_error_count Count error on publishing
# TYPE rabbitmq_publish_error_count counter
rabbitmq_publish_error_count{exchange="exchange",routing_key="key"} 1
```

## Usage

### Consumer

```go
consumerMetrics := rabbitmq_metrics.NewConsumerStorage(metrics.DefaultRegistry)
...
consumerMetrics.ObserveConsumeDuration("exchange", "key", duration)
consumerMetrics.IncSuccessCount("exchange", "key")
```

### Publisher

```go
publisherMetrics := rabbitmq_metrics.NewPublisherStorage(metrics.DefaultRegistry)
...
publisherMetrics.ObservePublishDuration("exchange", "key", duration)
publisherMetrics.IncPublishError("exchange", "key")
```

# Package `publisher`

Пакет `publisher` предоставляет реализацию паблишера Apache Kafka с поддержкой middleware, метрик и управления жизненным
циклом. Интегрируется с пакетом [`kafkax`](../.) для работы с Kafka.

## Types

### Publisher

Основная структура для отправки сообщений в Kafka. Поддерживает:

- Автоматическую подстановку топика сообщений
- Цепочки middleware
- Сбор метрик производительности
- Проверку состояния (healthcheck)

**Methods:**

#### `New(writer *kafka.Writer, topic string, metrics *Metrics, opts ...Option) *Publisher`

Создает новый продюсер из изкоуровневого райтера из библиотеки `kafka-go`.

Основные опции:

- `WithMiddlewares(mws ...Middleware) Option` – добавить middleware в цепочку обработки публикуемых сообщений.

#### `(p *Publisher) Publish(ctx context.Context, msgs ...kafka.Message) error`

Отправить сообщения в Kafka. Автоматически:

- Устанавливает топик сообщения, если не указана
- Применяет цепочку middleware
- Обновляет метрики

#### `(p *Publisher) Close() error`

Остановить паблишер, закрыв соединения.

#### `(p *Publisher) Healthcheck(ctx context.Context) error`

Проверить активность паблишера. Возвращает ошибку при проблемах с подключением.

### Metrics

Структура для сбора и отправки метрик продюсера в Prometheus.

**Methods:**

#### `NewMetrics(sendMetricPeriod time.Duration, writer *kafka.Writer, publisherId string) *Metrics`

Структура для сбора и отправки метрик паблишера в Prometheus.

#### `(m *Metrics) Send(stats kafka.WriterStats)`

Единожды отправить метрики.

#### `(m *Metrics) Run()`

Запустить периодическую отправку метрик.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/txix-open/isp-kit/kafkax"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	log2 "github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/retry"
)

const (
	maxRetryElapsedTime = 5 * time.Second
)

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test",
	})

	metrics := publisher.NewMetrics(
		10*time.Second,
		writer,
		"test-publisher",
	)

	publisher := publisher.New(
		writer,
		"test",
		metrics,
		publisher.WithMiddlewares(
			kafkax.PublisherLog(logger, true),
			/* publishing with retries */
			kafkax.PublisherRetry(retry.NewExponentialBackoff(maxRetryElapsedTime)),
		),
	)

	/* send message */
	err = publisher.Publish(context.Background(), kafka.Message{
		Key:   []byte("data"),
		Value: []byte(`{"status": "processed"}`),
	})
	if err != nil {
		log.Fatal(err)
	}
}

```
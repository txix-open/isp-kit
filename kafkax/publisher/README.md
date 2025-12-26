# Package `publisher`

Пакет `publisher` предоставляет реализацию паблишера Apache Kafka с поддержкой middleware, метрик и управления жизненным
циклом. Интегрируется с пакетом [`kafkax`](../.) для работы с Kafka.

## Types

### Publisher

Основная структура для отправки сообщений в Kafka. Поддерживает:

- Автоматическую подстановку топика сообщений
- Цепочки middleware
- Проверку состояния (healthcheck)

**Methods:**

#### `New(client *kgo.Client, topic string, opts ...Option) *Publisher`

Создает новый продюсер из клиента из библиотеки `franz-go`.

Основные опции:

- `WithMiddlewares(mws ...Middleware) Option` – добавить middleware в цепочку обработки публикуемых сообщений.

#### `(p *Publisher) Publish(ctx context.Context, rs ...*kgo.Record) error`

Отправить сообщения в Kafka. Автоматически:

- Устанавливает топик сообщения, если не указана
- Применяет цепочку middleware

#### `(p *Publisher) Close() error`

Остановить паблишер, закрыв соединения.

#### `(p *Publisher) Healthcheck(ctx context.Context) error`

Проверить активность паблишера. Возвращает ошибку при проблемах с подключением.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kprom"
	"github.com/txix-open/isp-kit/metrics"
	"log"
	"time"

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

	metricsTest := kprom.NewMetrics(
		"kafka_publisher",
	)

	defaultRegistry := metrics.DefaultRegistry
	defaultRegistry.GetOrRegister(metricsTest)

	client, err := kgo.NewClient(kgo.SeedBrokers("localhost:9092"), kgo.WithHooks(metricsTest))
	if err != nil {
		log.Fatal(err)
	}

	publisher := publisher.New(
		client,
		"test",
		publisher.WithMiddlewares(
			kafkax.PublisherLog(logger, true),
			/* publishing with retries */
			kafkax.PublisherRetry(retry.NewExponentialBackoff(maxRetryElapsedTime)),
		),
	)

	/* send message */
	err = publisher.Publish(context.Background(), &kgo.Record{
		Key:   []byte("data"),
		Value: []byte(`{"status": "processed"}`),
	})
	if err != nil {
		log.Fatal(err)
	}
}

```
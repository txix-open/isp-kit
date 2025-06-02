# Package `consumer`

Пакет `consumer` предоставляет реализацию консумера Apache Kafka с поддержкой параллельной обработки сообщений,
middleware, метрик и наблюдения за состоянием.
Интегрируется с пакетом [`kafkax`](../.) для полноценной работы с Kafka.

## Types

### Consumer

Основная структура для чтения и обработки сообщений из Kafka. Поддерживает:

- Параллельную обработку (конкурентность)
- Middleware-цепочки
- Сбор метрик

**Methods:**

#### `New(reader *kafka.Reader, handler Handler, concurrency int, metrics *Metrics, opts ...Option) *Consumer`

Создать нового консумера из низкоуровневого ридера из библиотеки `kafka-go` с указанным обработчиком сообщений
`Handler`.

Основные опции:

- `WithMiddlewares(mws ...Middleware) Option` – добавить middleware в цепочку обработки получаемых сообщений.
- `WithObserver(observer Observer) Option` – добавить реализацию интерфейса `Observer`.

#### `(c *Consumer) Run(ctx context.Context)`

Запустить чтение сообщений из Kafka и их обработку.

#### `(c *Consumer) Close() error`

Остановить консумера, завершить все активные обработки.

#### `(c *Consumer) Healthcheck(ctx context.Context) error`

Проверить активность консумера. Возвращает ошибку, если консумер не может получать сообщения.

### Delivery

Структура, представляющая полученное сообщение Kafka. Обеспечивает безопасное управление подтверждением (commit)
сообщения.

**Methods:**

#### `(d *Delivery) Commit(ctx context.Context) error`

Подтвердить успешную обработку сообщения. Должен вызываться только один раз за сообщение.

#### `(d *Delivery) Source() *kafka.Message`

Получить исходное сообщение Kafka (топик, партиция, ключ, значение).

#### `(d *Delivery) Done()`

Отметить завершение обработки (используется для синхронизации).

#### `(d *Delivery) ConsumerGroupId() string`

Получить groupId консумера.

### Metrics

Структура для сбора и отправки метрик консумера в Prometheus.

**Methods:**

#### `NewMetrics(sendMetricPeriod time.Duration, reader *kafka.Reader, consumerId string) *Metrics`

Создает сборщик метрик из низкоуровневого ридера Kafka.

#### `(m *Metrics) Send(stats kafka.ReaderStats)`

Единожды отправить метрики.

#### `(m *Metrics) Run()`

Запустить периодическую отправку метрик.

### LogObserver

Реализация интерфейса `consumer.Observer` для логирования событий консумера.

**Methods:**

#### `NewLogObserver(ctx context.Context, logger log.Logger) LogObserver`

Конструктор обсервера.

#### `(l LogObserver) ConsumerError(err error)`

Залогировать сообщение об ошибке консумера.

#### `(l LogObserver) BeginConsuming()`

Залогировать сообщение о начале получения данных от консумера.

#### `(l LogObserver) CloseStart()`

Залогировать сообщение о начале процесса завершения работы консумера.

#### `(l LogObserver) CloseDone()`

Залогировать сообщение об окончании процесса завершения работы консумера.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/kafkax"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	log2 "github.com/txix-open/isp-kit/log"
)

func noopHandlerFn(ctx context.Context, delivery *consumer.Delivery) {
	/* put here business logic */
	_ = delivery.Commit(ctx)
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   "test",
		GroupID: "test",
	})

	observer := consumer.NewLogObserver(context.Background(), logger)
	consumer := consumer.New(
		reader,
		consumer.HandlerFunc(noopHandlerFn),
		3,   /* concurrency */
		nil, /* metrics */
		consumer.WithMiddlewares(kafkax.ConsumerLog(logger, true)),
		consumer.WithObserver(observer),
	)

	consumer.Run(context.Background())
}

```
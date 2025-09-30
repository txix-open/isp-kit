# Package `grmqx`

Пакет `grmqx` предназначен для работы с брокером сообщений RabbitMQ, предоставляющий высокоуровневую абстракцию
над [grmq](https://github.com/txix-open/grmq) с дополнительными возможностями:

- Автоматическое объявление топологии
- Пакетная обработка сообщений
- Интеграция с метриками и трейсингом
- Гибкая система повторов и автогенерация DLQ
- Контекстное логирование
- Поддержка аргументов очередей (x-single-active-consumer и др.)

## Types

### Client

Структура `Client` управляет подключением к RabbitMQ и жизненным циклом обработчиков.

**Methods:**

#### `New(logger log.Logger) *Client`

Конструктор клиента RabbitMQ.

#### `(c *Client) Upgrade(ctx context.Context, config Config) error`

Обновить конфигурацию, синхронно инициализировать клиент с гарантией готовности всех компонентов:

- Блокировка и ожидание первого успешно установленной сессии.
- Запуск всех консумеров, инициализация всех паблишеров и применение всех объявлений.
- Вернет первую возникшую ошибку во время открытия первой сессии или `nil`

#### `(c *Client) UpgradeAndServe(ctx context.Context, config Config)`

Аналогично методу `Upgrade` обновляет конфигурацию, но не ждет первой успешно установленной сессии. Передает в Observer
возникшие ошибки и ретраит.

#### `(c *Client) Healthcheck(ctx context.Context) error`

Проверить возможность подключения к брокеру.

#### `(c *Client) Close()`

Закрыть соединения и остановить клиент.

### Connection

Конфигурация параметров подключения.

**Methods:**

#### `(c Connection) Url() string`

Получить URL подключения к RabbitMQ.

### Publisher

Конфигурация параметров паблишера.

**Methods:**

#### `(p Publisher) DefaultPublisher(restMiddlewares ...publisher.Middleware) *publisher.Publisher`

Создать паблишера с предустановленными middleware и настройками:

- PersistentMode.
- Генерация и добавление в заголовки requestId.
- Метрики и трейсинг.

Опциональные middleware:

- `PublisherLog(logger log.Logger, logBody bool) publisher.Middleware` – логировать публикуемые сообщения.
- `PublisherRequestId() publisher.Middleware` – генерация и добавление в заголовки requestId (установленно
  по-умолчанию).
- `PublisherRetry(retrier Retrier) publisher.Middleware` – добавить ретраи при возникновении ошибок публикации при
  помощи объекта, реализующего интерфейс `Retrier`.
- `PublisherMetrics(storage PublisherMetricStorage) publisher.Middleware` – добавить метрики при помощи объекта,
  реализующего интерфейс `PublisherMetricStorage` (установленно по-умолчанию).

### Consumer

Конфигурация параметров консумера.

**Methods:**

#### `(c Consumer) DefaultConsumer(handler consumer.Handler, restMiddlewares ...consumer.Middleware) consumer.Consumer`

Создать консумера с обработчиком сообщений, реализующим интерфейс `consumer.Handler`, и базовыми настройками.

Опциональные middleware:

- `ConsumerLog(logger log.Logger, logBody bool) consumer.Middleware` – логирование информации о получаемых сообщениях;
  можно включить/выключить логирование тела сообщения.
- `ConsumerRequestId() consumer.Middleware` – получить requestId из заголовка и сохранить его в контексте (установленно
  по-умолчанию).

### BatchConsumer

Конфигурация параметров батч-консумера.

**Methods:**

#### `(b BatchConsumer) ConsumerConfig() Consumer`

Конвертирует `BatchConsumer` в стандартную конфигурацию `Consumer`. Фиксирует `Concurrency = 1` и наследует все
остальные параметры.

####

`(b BatchConsumer) DefaultConsumer(handler batch_handler.BatchHandlerAdapter, restMiddlewares ...consumer.Middleware) consumer.Consumer`

Создать батч-консумера с пакетной обработкой сообщений. Обработчик сообщений должен реализовывать интерфейс
`batch_handler.BatchHandlerAdapter` или быть преобразованным к `batch_handler.BatchHandlerAdapterFunc`, 
если это функция-обработчик.

### LogObserver

Реализация интерфейса `grmq.Observer` для логирования событий RabbitMQ-клиента.

**Methods:**

#### `NewLogObserver(ctx context.Context, logger log.Logger) LogObserver`

Конструктор обсервера.

#### `(l LogObserver) ClientReady()`

Залогировать сообщение о готовности RabbitMQ-клиента.

#### `(l LogObserver) ClientError(err error)`

Залогировать сообщение об ошибке RabbitMQ-клиента.

#### `(l LogObserver) ConsumerError(consumer consumer.Consumer, err error)`

Залогировать сообщение об ошибке консумера.

#### `(l LogObserver) PublisherError(publisher *publisher.Publisher, err error)`

Залогировать сообщение об ошибке паблишера.

#### `(l LogObserver) ShutdownStarted()`

Залогировать сообщение о начале процесса завершения работы RabbitMQ-клиента.

#### `(l LogObserver) ShutdownDone()`

Залогировать сообщение об окончании процесса завершения работы RabbitMQ-клиента.

#### `(l LogObserver) PublishingFlow(publisher *publisher.Publisher, flow bool)`

Залогировать сообщение с информацией о потоке публикации.

#### `(l LogObserver) ConnectionBlocked(reason string)`

Залогировать сообщение о блокировке соединения с указанием причины.

#### `(l LogObserver) ConnectionUnblocked()`

Залогировать сообщение о разблокировке соединения.

## Functions

#### `TopologyFromConsumers(consumers ...Consumer) topology.Declarations`

Сгенерировать декларации топологии RabbitMQ на основе конфигураций консумеров.

#### `JoinDeclarations(declarations ...topology.Declarations) topology.Declarations`

Объединить несколько деклараций топологии в одну..

#### `NewResultHandler(logger log.Logger, adapter handler.SyncHandlerAdapter) handler.Sync`

Создает готовый синхронный обработчик сообщений RabbitMQ с предустановленными инструментами для:

- Логирования
- Сбора метрик
- Трейсинга
- Восстановления при панике

#### `NewResultBatchHandler(logger log.Logger, adapter batch_handler.SyncHandlerAdapter) batch_handler.Sync`

Создает готовый синхронный пакетный обработчик сообщений RabbitMQ с предустановленными инструментами для:

- Логирования
- Сбора метрик
- Восстановления при панике

## Usage

### Consumer & publisher

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/grmqx/handler"
	log2 "github.com/txix-open/isp-kit/log"
)

type noopHandler struct{}

func (h noopHandler) Handle(ctx context.Context, delivery *consumer.Delivery) handler.Result {
	/* put here business logic */
	return handler.Ack()
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	rmqCli := grmqx.New(logger)
	conn := grmqx.Connection{
		Host:     "test",
		Port:     5672,
		Username: "test",
		Password: "test",
		Vhost:    "/",
	}
	publisherCfg := grmqx.Publisher{
		Exchange:   "",
		RoutingKey: "queue-2",
	}
	consumerCfg := grmqx.Consumer{Queue: "queue"}

	/* create consumer & publisher from configs */
	consumer := consumerCfg.DefaultConsumer(
		grmqx.NewResultHandler(logger, noopHandler{}),
		grmqx.ConsumerLog(logger, true),
	)
	publisher := publisherCfg.DefaultPublisher()
	err = rmqCli.Upgrade(context.Background(), grmqx.NewConfig(
		conn.Url(),
		grmqx.WithConsumers(consumer),
		grmqx.WithPublishers(publisher),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(consumerCfg)),
	))
	if err != nil {
		log.Fatal(err)
	}
}

```

### Batch consumer

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/grmqx/batch_handler"
	log2 "github.com/txix-open/isp-kit/log"
)

type batchHandler struct{}

func (h batchHandler) Handle(batch []batch_handler.Item) *batch_handler.Result {
	result := batch_handler.NewResult()
	/* put here business logic */
	for idx := range batch {
		result.AddAck(idx)
	}
	return result
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	rmqCli := grmqx.New(logger)
	conn := grmqx.Connection{
		Host:     "test",
		Port:     5672,
		Username: "test",
		Password: "test",
		Vhost:    "/",
	}

	consumerCfg := grmqx.BatchConsumer{
		Queue:             "queue-1",
		BatchSize:         100,
		PurgeIntervalInMs: 60000,
	}
	consumer := consumerCfg.DefaultConsumer(
		batchHandler{},
		grmqx.ConsumerLog(logger, true),
	)

	err = rmqCli.Upgrade(context.Background(), grmqx.NewConfig(
		conn.Url(),
		grmqx.WithConsumers(consumer),
		grmqx.WithDeclarations(grmqx.TopologyFromConsumers(
			consumerCfg.ConsumerConfig(),
		)),
	))
	if err != nil {
		log.Fatal(err)
	}
}

```
# Package `kafkax`

Пакет `kafkax` предоставляет высокоуровневую абстракцию для работы с Apache Kafka, включая продюсеры, консьюмеры,
аутентификацию, TLS, middleware для логирования, метрик и обработки ошибок. Поддерживает динамическое обновление
конфигурации без остановки сервиса.

## Types

### Client

Центральный клиент для управления продюсерами и консьюмерами Kafka.

**Methods:**

#### `New(logger log.Logger) *Client`

Создать новый клиент с логгером.

#### `(c *Client) UpgradeAndServe(ctx context.Context, config Config)`

Обновить конфигурацию и перезапустить подключения:

- Останавливает старые соединения
- Инициализирует новые продюсеры/консьюмеры
- Запускает обработку сообщений

#### `(c *Client) Healthcheck(ctx context.Context) error`

Проверить доступность всех продюсеров и консьюмеров.

#### `(c *Client) Close()`

Остановить все соединения.

### PublisherConfig

Конфигурация паблишера для отправки сообщений.

**Methods:**

####

`DefaultPublisher(logCtx context.Context, logger log.Logger, restMiddlewares ...publisher.Middleware) *publisher.Publisher`

Создать паблишера с настройками по умолчанию:

- Таймаут отправки: 10 сек
- Размер батча: 64 МБ
- Middleware для метрик и requestId

### ConsumerConfig

Конфигурация консумера для чтения сообщений.

**Methods:**

####

`DefaultConsumer(logCtx context.Context, logger log.Logger, handler consumer.Handler, restMiddlewares ...consumer.Middleware) consumer.Consumer`

Создать консьюмер с настройками по умолчанию:

- Таймаут подключения: 5 сек
- Интервал коммита: 1 сек
- Middleware для логирования и requestId

### LogObserver

Реализация интерфейса `kafkax.Observer` для логирования событий Kafka-клиента.

**Methods:**

#### `NewLogObserver(ctx context.Context, logger log.Logger) LogObserver`

Конструктор обсервера.

#### `(l LogObserver) ClientReady()`

Залогировать сообщение о готовности Kafka-клиента.

#### `(l LogObserver) ClientError(err error)`

Залогировать сообщение об ошибке Kafka-клиента.

#### `(l LogObserver) ShutdownStarted()`

Залогировать сообщение о начале процесса завершения работы Kafka-клиента.

#### `(l LogObserver) ShutdownDone()`

Залогировать сообщение об окончании процесса завершения работы Kafka-клиента.

## Functions

#### `NewResultHandler(logger log.Logger, adapter handler.SyncHandlerAdapter) handler.Sync`

Создать обработчик сообщений с:

- Логированием
- Метриками
- Поддержкой синхронной обработки
- Восстановлением при панике

#### `PublisherLog(logger log.Logger, logBody bool) publisher.Middleware`

Middleware для логирования информации о публикуемых сообщениях. Логирует тело сообщения, если `logBody = true`

#### `ConsumerLog(logger log.Logger, logBody bool) consumer.Middleware`

Middleware для логирования информации о получаемых сообщениях. Логирует тело сообщения, если `logBody = true`

#### `PublisherRetry(retrier Retrier) publisher.Middleware`

Middleware для повторной отправки сообщений при ошибках. Принимает реализацию интерфейса `Retrier` с логикой выдержки
ретрая.

#### `PublisherMetrics(storage PublisherMetricStorage) publisher.Middleware`

Middleware для сбора следующих метрик паблишера:

- Время публикации сообщений
- Размеры публикуемых сообщений
- Количество ошибок

#### `PublisherRequestId() publisher.Middleware`

Middleware, добавляющая в заголовки сообщений паблишера requestId из контекста. Автоматически генерирует requestId, если
в контексте его нет.

#### `ConsumerRequestId() consumer.Middleware`

Middleware, добавляющая в контекст requestId из заголовков полученных сообщений. Автоматически генерирует requestId,
если в заголовках его нет.

## Usage

### Consumer & publisher

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/kafka/handler"
	"github.com/txix-open/isp-kit/kafkax"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	log2 "github.com/txix-open/isp-kit/log"
)

type noopHandler struct{}

func (h noopHandler) Handle(ctx context.Context, d *consumer.Delivery) handler.Result {
	/* put here business logic */
	return handler.Commit()
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	publisherCfg := kafkax.PublisherConfig{
		Addresses:             []string{"localhost:9092"},
		Topic:                 "topic-2",
		BatchSizePerPartition: 1,
		Auth: &kafkax.Auth{
			Username: "test",
			Password: "test",
		},
	}
	consumerCfg := kafkax.ConsumerConfig{
		Addresses:   []string{"localhost:9092"},
		Topic:       "topic-1",
		GroupId:     "test",
		Concurrency: 1,
		Auth: &kafkax.Auth{
			Username: "test",
			Password: "test",
		},
	}

	consumer := consumerCfg.DefaultConsumer(
		context.Background(),
		logger,
		handler.NewSync(logger, noopHandler{}),
	)
	publisher := publisherCfg.DefaultPublisher(context.Background(), logger)

	cli := kafkax.New(logger)
	cli.UpgradeAndServe(context.Background(), kafkax.NewConfig(
		kafkax.WithConsumers(consumer),
		kafkax.WithPublishers(publisher),
	))
}

```
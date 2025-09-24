# Package `stompx`

Пакет `stompx` предоставляет высокоуровневую обёртку над STOMP-протоколом, реализуя удобные инструменты для создания потребителей и издателей сообщений, с поддержкой middleware, логирования, повторных попыток и управления группой потребителей и издателей.

## Types

### Config

Конфигурация stompx-клиента.

**Fields:**

#### `Consumers  []*consumer.Watcher`

Массив потребителей.

#### `Publishers []*publisher.Publisher`

Массив издателей.

**Methods:**

#### `NewConfig(opts ...ConfigOption) Config`

Создаёт конфигурацию клиента с указанными опциями.

### ConfigOption

Функция, применяющая опции к `Config`.

**Functions:**

#### `WithConsumers(consumers ...*consumer.Watcher) ConfigOption`

Добавляет клиенты потребителей в конфигурацию клиента.

#### `WithPublishers(publishers ...*publisher.Publisher) ConfigOption`

Добавляет клиенты издателей в конфигурацию клиента.

### ConsumerConfig

Конфигурация потребителя сообщений.

**Fields:**

#### `Address string`

Адрес брокера (обязательное).

#### `Queue string`

Имя очереди (обязательное).

#### `Concurrency int`

Количество обработчиков (по умолчанию 1).

#### `PrefetchCount int`

Количество предзагруженных сообщений.

#### `Username string`

Имя пользователя.

#### `Password string`

Пароль.

#### `ConnHeaders  map[string]string`

Дополнительные заголовки подключения.

### PublisherConfig

Конфигурация издателя сообщений.

**Fields:**

#### `Address string`

Адрес брокера (обязательное).

#### `Queue string`

Имя очереди (обязательное).

#### `Username string`

Имя пользователя.

#### `Password string`

Пароль.

#### `ConnHeaders map[string]string`

Дополнительные заголовки подключения.

### Client

Клиент, включающий в себя группу потребителей и издателей, способный обновлять соединения и перезапускаться при изменении конфигурации.

**Methods:**

#### `New(logger log.Logger) *Client`

Создать новый клиент с логгером.

#### `(g *Client) Upgrade(ctx context.Context, config Config) error`

Обновить конфигурацию, синхронно инициализировать клиент с гарантией готовности всех компонентов:

- Блокировка и ожидание первой успешно установленной сессии.
- Запуск всех потребителей, инициализация всех издателей.
- Вернет первую возникшую ошибку во время открытия первой сессии или `nil`.

#### `(g *Client) UpgradeAndServe(ctx context.Context, config Config) error`

Обновить конфигурацию и перезапустить подключения:

- Останавливает старые соединения,
- Инициализирует новые потребители/издатели,
- Запускает обработку сообщений потребителями.

#### `(g *Client) Close() error`

Завершает все активные подключения.

### LogObserver

Наблюдатель за событиями жизненного цикла потребителя (ошибки, запуск, остановка).

**Methods:**

#### `NewLogObserver(logger log.Logger) LogObserver`

Создаёт наблюдателя за событиями с указанным логером.

#### `Error(c Consumer, err error)`

Логирует ошибку потребителя.

#### `CloseStart(c Consumer)` / `CloseDone(c Consumer)`

Логируют процесс остановки.

#### `BeginConsuming(c Consumer)`

Логирует начало потребления сообщений.

## Functions

### `DefaultConsumer(cfg ConsumerConfig, handler consumer.Handler, logger log.Logger, restMiddlewares ...consumer.Middleware) consumer.Config`

Создаёт конфигурацию потребителя с поддержкой логирования, middleware и подключения по заданным параметрам.

### `DefaultPublisher(cfg PublisherConfig, restMiddlewares ...publisher.Middleware) *publisher.Publisher`

Создаёт издателя сообщений с middleware и настройками подключения.

### `NewResultHandler(logger log.Logger, adapter handler.HandlerAdapter) handler.ResultHandler`

Создаёт обработчик результата с логированием и восстановлением сервиса при панике.

## Middleware

### PublisherPersistent

Добавляет заголовок `persistent=true` ко всем исходящим сообщениям.

### PublisherLog

Логирует отправку сообщений.

### PublisherRequestId

Добавляет Request-Id в заголовки сообщений.

### PublisherRetry

Повторяет публикацию при ошибках с использованием заданного `Retrier`.

### ConsumerLog

Логирует входящие сообщения.

### ConsumerRequestId

Извлекает или генерирует Request-Id и сохраняет его в контексте запроса.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/shutdown"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/requestid"
	"github.com/txix-open/isp-kit/stompx"
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/publisher"
)

func main() {
	logger := log.New()

	// Создаём обработчик сообщений
	handler := stompx.NewResultHandler(logger, stompx.HandlerFunc(func(ctx context.Context, msg []byte) error {
		logger.Info("message received", log.StringType("body", string(msg)))
		return nil
	}))

	// Конфигурация потребителя
	consumerCfg := stompx.ConsumerConfig{
		Address:  "tcp://localhost:61613",
		Queue:    "/queue/example",
		Username: "admin",
		Password: "admin",
	}

	publisherCfg := stompx.PublisherConfig{
		Address:  "tcp://localhost:61613",
		Queue:    "/queue/example",
		Username: "admin",
		Password: "admin",
	}

	consumerCli := consumer.NewWatcher(stompx.DefaultConsumer(consumerCfg, handler, logger))
	publisherCli := stompx.DefaultPublisher(publisherCfg)

	// Создаём конфигурацию
	config := stompx.NewConfig(
		stompx.WithConsumers(consumerCli),
		stompx.WithPublishers(publisherCli))

	// Создаём клиент
	сli := stompx.NewClient(logger)

	// Обработка завершения приложения
	shutdown.On(func() {
		logger.Info("shutting down...")
		_ = сli.Close()
		logger.Info("shutdown completed")
	})

	// Запускаем клиент
	сli.UpgradeAndServe(context.Background(), config)
	...
}
```

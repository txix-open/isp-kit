# Package `stompx`

Пакет `stompx` предоставляет высокоуровневую обёртку над STOMP-протоколом, реализуя удобные инструменты для создания потребителей и издателей сообщений, с поддержкой middleware, логирования, повторных попыток и управления группой потребителей.

## Types

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

#### `ConnHeaders  map[string]string`

Дополнительные заголовки подключения.

### ConsumerGroup

Группа потребителей, способная обновляться и перезапускаться при изменении конфигурации.

**Methods:**

#### `(g *ConsumerGroup) Upgrade(ctx context.Context, consumers ...Consumer) error`

Применяет новую конфигурацию без запуска.

#### `(g *ConsumerGroup) UpgradeAndServe(ctx context.Context, consumers ...Consumer) error`

Применяет и запускает новых потребителей.

#### `(g *ConsumerGroup) Close() error`

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

Создаёт обработчик результата с логированием.

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

	consumer, err := stompx.DefaultConsumer(consumerCfg, handler, logger)
	if err != nil {
		log.Fatal(err)
	}

	// Создаём группу потребителей
	group := stompx.NewConsumerGroup(logger)

	// Обработка завершения приложения
	shutdown.On(func() {
		logger.Info("shutting down...")
		_ = group.Close()
		logger.Info("shutdown completed")
	})

	// Запускаем потребителей
	err = group.UpgradeAndServe(context.Background(), consumer)
	if err != nil {
		logger.Fatal("failed to start consumer group", log.String("error", err.Error()))
	}
	...
}
```

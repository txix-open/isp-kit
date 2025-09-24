# Package `publisher`

Пакет `publisher` предоставляет удобный механизм отправки сообщений через STOMP-протокол с поддержкой middleware и повторного подключения.

## Types

### Publisher

Структура `Publisher` инкапсулирует логику подключения к STOMP-брокеру и отправки сообщений с возможностью настройки middleware и опций соединения.

**Methods:**

#### `NewPublisher(address string, queue string, opts ...Option) *Publisher`

Создаёт новый экземпляр `Publisher` с заданным адресом брокера, очередью и опциями.

#### `(p *Publisher) Publish(ctx context.Context, msg *Message) error`

Публикует сообщение в очередь, указанную при создании.

#### `(p *Publisher) PublishTo(ctx context.Context, queue string, msg *Message) error`

Публикует сообщение в указанную очередь.

#### `(p *Publisher) Close() error`

Закрывает текущее STOMP-соединение.

#### `(p *Publisher) Healthcheck(ctx context.Context) error`

Проверить работоспособность соединения паблишера. Возвращает ошибку при проблемах с подключением.

### Message

Структура `Message` представляет сообщение для публикации.

**Methods:**

#### `Json(body []byte) *Message`

Создаёт сообщение с типом `application/json`.

#### `PlainText(body []byte) *Message`

Создаёт сообщение с типом `plain/text`.

#### `(m *Message) WithHeader(key string, value string) *Message`

Добавляет заголовок к сообщению.

### Option

Функция, применяющая опции к `Publisher`.

**Functions:**

#### `WithMiddlewares(mws ...Middleware) Option`

Добавляет middleware в цепочку обработки публикации.

#### `WithConnectionOptions(connOpts ...consumer.ConnOption) Option`

Передаёт опции соединения для подключения к STOMP-брокеру.

### Middleware

Функция, принимающая `RoundTripper` и возвращающая новый `RoundTripper`. Позволяет оборачивать логику публикации.

### RoundTripper

Интерфейс, реализующий отправку сообщения.

### RoundTripperFunc

Функциональный адаптер, реализующий интерфейс `RoundTripper`.

**Methods:**

#### `(f RoundTripperFunc) Publish(ctx context.Context, queue string, msg *Message) error`

Вызывает саму функцию.

### PublishOption

Тип `PublishOption` представляет собой функцию, применяющую изменения к STOMP-фрейму. Полностью совместим с опциями `github.com/go-stomp/stomp/v3/frame`.

## Usage

### Basic publisher usage

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/stompx/publisher"
)

func main() {
	// Создание нового паблишера
	pub := publisher.NewPublisher("localhost:61613", "/queue/example")

	// Подготовка сообщения
	msg := publisher.Json([]byte(`{"event":"user_created"}`)).
		WithHeader("X-Custom-Header", "value")

	// Публикация сообщения
	err := pub.Publish(context.Background(), msg)
	if err != nil {
		log.Fatalf("publish error: %v", err)
	}

	// Завершение соединения
	err = pub.Close()
	if err != nil {
		log.Fatalf("close error: %v", err)
	}
}
```

### With middlewares and connection options

```go
import (
	"github.com/go-stomp/stomp/v3"
	"github.com/txix-open/isp-kit/stompx/publisher"
)

pub := publisher.NewPublisher(
	"localhost:61613",
	"/queue/another",
	publisher.WithConnectionOptions(stomp.ConnOpt.Login("user", "pass")),
	publisher.WithMiddlewares(
		func(next publisher.RoundTripper) publisher.RoundTripper {
			return publisher.RoundTripperFunc(func(ctx context.Context, queue string, msg *publisher.Message) error {
				// логика до
				err := next.Publish(ctx, queue, msg)
				// логика после
				return err
			})
		},
	),
)
```

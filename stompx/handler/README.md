# Package `handler`

Пакет `handler` предоставляет инструменты для обработки входящих STOMP-сообщений с использованием middleware и результатной семантики (`Ack` / `Requeue`).

## Types

### Result

Структура `Result` определяет, каким образом обрабатывать входящее сообщение после выполнения логики хендлера.

**Fields:**

#### `Ack bool`

Сообщение успешно обработано и будет подтверждено (`ACK`).

#### `Requeue bool`

Сообщение не обработано, требуется повторная доставка (`NACK`).

#### `Err error`

Ошибка, связанная с повторной доставкой.

**Methods:**

#### `Ack() Result`

Создаёт результат, означающий успешную обработку сообщения (будет отправлен `ACK`).

#### `Requeue(err error) Result`

Создаёт результат, означающий необходимость повторной доставки сообщения (будет отправлен `NACK`).

### HandlerAdapter

Интерфейс `HandlerAdapter` определяет контракт для обработки STOMP-сообщений.

### AdapterFunc

Адаптер для использования обычной функции в качестве `HandlerAdapter`.

**Methods:**

#### `(a AdapterFunc) Handle(ctx, msg)`

Вызов соответствующей функции и возврат результата обработки.

### ResultHandler

Структура `ResultHandler` реализует конечную обработку сообщения, включая выполнение middleware-цепочки, вызов хендлера и вызов `Ack()` / `Nack()` у delivery в зависимости от результата.

**Methods:**

#### `NewHandler(logger log.Logger, adapter HandlerAdapter, middlewares ...Middleware) ResultHandler`

Создаёт экземпляр `ResultHandler`, оборачивая адаптер через указанные middleware.

#### `(r ResultHandler) Handle(ctx context.Context, delivery *consumer.Delivery)`

Вызывает адаптер и в зависимости от результата вызывает `Ack()` или `Nack()` у delivery. Ошибки логируются.

## Middleware

Middleware — функция, принимающая `HandlerAdapter` и возвращающая обёрнутый `HandlerAdapter`.

### Log

Логирует действия хендлера (ACK или Requeue) с указанием destination.

### Recovery

Предотвращает падение сервиса при панике в обработчике, преобразуя ее в ошибку.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"errors"
	"github.com/go-stomp/stomp/v3"
	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/handler"
)

func main() {
	logger := log.New(log.Config{})

	h := handler.NewHandler(logger, handler.AdapterFunc(func(ctx context.Context, msg *stomp.Message) handler.Result {
		if string(msg.Body) == "fail" {
			return handler.Requeue(errors.New("failed to handle message"))
		}
		return handler.Ack()
	}), handler.Log(logger))

	// Эмуляция входящего сообщения
	delivery := &consumer.DeliveryMock{
		Message: &stomp.Message{
			Body:        []byte("hello"),
			Destination: "/queue/example",
		},
	}

	h.Handle(context.Background(), delivery)
}
```

### DeliveryMock (пример тестового использования)

```go
package consumer

import "github.com/go-stomp/stomp/v3"

// DeliveryMock используется в тестах для имитации доставки сообщений.
type DeliveryMock struct {
	Message *stomp.Message
	Acked   bool
	Nacked  bool
}

func (d *DeliveryMock) Source() *stomp.Message {
	return d.Message
}

func (d *DeliveryMock) Ack() error {
	d.Acked = true
	return nil
}

func (d *DeliveryMock) Nack() error {
	d.Nacked = true
	return nil
}
```

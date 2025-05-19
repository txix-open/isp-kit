# Package `consumer_tracing`

Пакет `consumer_tracing` добавляет поддержку трассировки (OpenTelemetry) для потребителей сообщений RabbitMQ, работающих через библиотеку `grmq/consumer`.

## Types

### Config

Структура `Config` описывает настройки трассировки для входящих сообщений.

**Fields:**

#### `Provider tracing.TracerProvider`

Провайдер трассировки (по умолчанию `tracing.DefaultProvider`).

#### `Propagator tracing.Propagator`

Пропагатор контекста (по умолчанию `tracing.DefaultPropagator`).

**Methods:**

#### `NewConfig() Config`

Создаёт конфигурацию трассировки с провайдером и пропагатором по умолчанию.

#### `(c Config) Middleware() handler.Middleware`

Возвращает middleware, который:

- Извлекает контекст трассировки из заголовков сообщения.
- Создаёт новый span с данными о сообщении (exchange, routing key, операция "deliver").
- Завершает span со статусом в зависимости от результата обработки:

  - `Ack`: Успешно — `StatusCode=OK`
  - `Requeue`, `Retry`, `MoveToDlq`: Ошибка записывается и выставляется статус `Error`

Если трассировка отключена (noop), middleware просто вызывает следующий обработчик.

## Usage

### Default usage flow

```go
package main

import (
	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/grmqx/handler"
	"github.com/txix-open/isp-kit/observability/tracing/consumer_tracing"
)

func main() {
	cfg := consumer_tracing.NewConfig()
	h := handler.NewSyncHandlerAdapter(myHandlerFunc)

	h = cfg.Middleware()(h)
	consumer.RegisterHandler("queue", h)
}
```

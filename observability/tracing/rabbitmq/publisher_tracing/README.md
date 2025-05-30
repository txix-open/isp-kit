# Package `publisher_tracing`

Пакет `publisher_tracing` предназначен для добавления поддержки трассировки (OpenTelemetry) в публикации сообщений RabbitMQ через библиотеку `grmq/publisher`.

## Types

### Config

Структура `Config` описывает настройки трассировки публикаций в RabbitMQ.

**Fields:**

#### `Provider tracing.TracerProvider`

Провайдер трассировки (по умолчанию `tracing.DefaultProvider`).

#### `Propagator tracing.Propagator`

Пропагатор контекста (по умолчанию `tracing.DefaultPropagator`).

**Methods:**

#### `NewConfig() Config`

Создаёт конфигурацию трассировки с провайдером и пропагатором по умолчанию.

#### `(c Config) Middleware() publisher.Middleware`

Возвращает middleware для трассировки публикаций в RabbitMQ. Если трассировка отключена (`noop`), возвращается пустой middleware без логики трассировки.

В случае активной трассировки middleware:

- Создаёт span с информацией о публикуемом сообщении (exchange, routing key).
- Вставляет в заголовки сообщения трассировочные идентификаторы.
- Завершает span после публикации, записывая статус и возможную ошибку.

## Usage

### Default usage flow

```go
package main

import (
	"github.com/txix-open/grmq/publisher"
	"github.com/txix-open/isp-kit/observability/tracing/publisher_tracing"
)

func main() {
	cfg := publisher_tracing.NewConfig()
	pub := publisher.New(
		publisher.WithMiddleware(cfg.Middleware()),
	)
	// теперь все публикации будут трассироваться
}
```

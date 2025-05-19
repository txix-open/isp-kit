# Package `client_tracing`

Пакет `client_tracing` предназначен для добавления middleware к gRPC-клиенту с поддержкой распределённой трассировки через OpenTelemetry. Позволяет автоматически создавать и передавать span-контекст при выполнении gRPC-запросов, а также записывать статусы и ошибки вызовов в трассировку.

## Types

### Config

Структура `Config` представляет конфигурацию трассировки для gRPC-клиента.

#### `func NewConfig() Config`

Создаёт конфигурацию трассировки с провайдером и пропагатором по умолчанию.

**Fields:**

#### `Provider`

Реализация интерфейса `tracing.TracerProvider`, используемая для создания спанов.

#### `Propagator`

Реализация интерфейса `tracing.Propagator`, используемая для инъекции и извлечения контекста трассировки.

**Methods:**

#### `func NewConfig() Config`

Создаёт конфигурацию трассировки с `provider` и `propagator` по умолчанию.

#### `func (c Config) Middleware() request.Middleware`

Возвращает middleware для клиента gRPC, реализующего трассировку запроса. Если трассировка отключена (noop-провайдер), возвращается middleware-заглушка.

Middleware:

- Создаёт новый спан для каждого gRPC-вызова с именем `GRPC call <endpoint>`.
- Устанавливает `span kind` в значение `Client`.
- Добавляет `request_id` в атрибуты спана.
- Инъецирует контекст в metadata запроса.
- В случае ошибки записывает её в спан и устанавливает статус `Error`.

## Usage

### Default usage flow

```go
package main

import (
    "github.com/txix-open/isp-kit/grpc/client"
    "github.com/txix-open/isp-kit/observability/tracing/grpc/client_tracing"
)
func main() {
    ...
    cfg := client_tracing.NewConfig()
    client := client.New(
        client.WithMiddleware(
            cfg.Middleware(),
        ),
    )
    ...
}
```

Для корректной работы необходимо, чтобы в контексте присутствовал `request_id`, а также был корректно настроен глобальный TracerProvider и Propagator OpenTelemetry.

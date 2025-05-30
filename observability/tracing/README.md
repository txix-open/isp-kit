# Package `tracing`

Пакет `tracing` предоставляет интерфейсы и реализацию для инициализации трассировки с помощью OpenTelemetry, включая поддержку `OTLP` экспортёра, `TracerProvider`, `Propagator`, а также noop-реализацию.

## Types

### Config

Конфигурация для инициализации трассировки:

**Fields:**

#### `Enable bool`

Включает или отключает трассировку.

#### `Address string`

Адрес OTLP-экспортера (например, `localhost:4318`).

#### `ModuleName string`

Название сервиса.

#### `ModuleVersion string`

Версия сервиса

#### `Environment string`

Окружение выполнения (например, `prod`, `staging`).

#### `InstanceId string`

Уникальный идентификатор инстанса.

#### `Attributes map[string]string`

Дополнительные атрибуты, которые будут прикреплены к каждому спану.

### TracerProvider

Псевдоним для стандартного интерфейса OpenTelemetry `TracerProvider`.

### Propagator

Псевдоним для интерфейса `TextMapPropagator`, используемого для распространения контекста трассировки между сервисами.

### Provider

Интерфейс трассировщика, поддерживающий корректное завершение сессии через `Shutdown()`.

**Methods:**
#### `func NewProviderFromConfiguration(ctx context.Context, logger log.Logger, config Config) (Provider, error)`

Создаёт и возвращает `TracerProvider` на основе переданной конфигурации. Возвращает `NoopProvider`, если `Enable == false`.

Использует OTLP экспортёр через HTTP и устанавливает атрибуты ресурса:

- окружение,
- имя сервиса,
- версия,
- идентификатор инстанса,
- пользовательские атрибуты.

## Constants

```go
const RequestId = attribute.Key("app.request_id")
```

Ключ атрибута, используемый для добавления request-id в спан. Применяется как часть метаданных запроса.

## Global variables

```go
var (
	DefaultPropagator Propagator     = propagation.TraceContext{}
	DefaultProvider   TracerProvider = NewNoopProvider()
)
```

Значения по умолчанию:

- `DefaultPropagator` — TraceContext propagator.
- `DefaultProvider` — noop-реализация трассировщика.

## NoopProvider

Реализация `TracerProvider`, которая не делает ничего (используется по умолчанию при отключенной трассировке).

**Methods:**

#### `func NewNoopProvider() NoopProvider`

Создаёт новый `NoopProvider`.

#### `(n NoopProvider) Tracer(name string, options ...trace.TracerOption) trace.Tracer`

Возвращает пустой `noop.Tracer`.

#### `(n NoopProvider) Shutdown(ctx context.Context) error`

Операция завершения, не выполняющая действий.

#### `func IsNoop(provider TracerProvider) bool`

Проверяет, является ли провайдер noop-реализацией.

## Usage

Пример настройки трассировки с использованием OTLP экспортёра и установкой глобального провайдера.
Завершение трассировки выполняется через `Shutdown`.

```go
package main

import (
    ...
    "github.com/txix-open/observability/tracing"
    "go.opentelemetry.io/otel"
    ...
)

func main(){
...
cfg := tracing.Config{
	Enable:        true,
	Address:       "localhost:4318",
	ModuleName:    "user-service",
	ModuleVersion: "1.0.0",
	Environment:   "production",
	InstanceId:    "abc123",
	Attributes: map[string]string{
		"region": "eu-central-1",
	},
}

provider, err := tracing.NewProviderFromConfiguration(context.Background(), logger, cfg)
if err != nil {
	log.Fatal(err)
}

otel.SetTracerProvider(provider)
defer provider.Shutdown(context.Background())
...
}
```

# Package `server_tracing`

Пакет `server_tracing` предоставляет middleware для серверов gRPC, реализующих трассировку входящих запросов с использованием OpenTelemetry.

## Types

### Config

Структура `Config` предназначена для конфигурации middleware трассировки gRPC-сервера. Она содержит провайдер трассировки и пропагатор контекста.

**Fields:**

#### `Provider tracing.TracerProvider`

Реализация интерфейса `tracing.TracerProvider`, используемая для создания спанов.

#### `Propagator tracing.Propagator`

Реализация интерфейса `tracing.Propagator`, используемая для инъекции и извлечения контекста трассировки.

**Methods:**

#### `func NewConfig() Config`

Создаёт конфигурацию трассировки с `provider` и `propagator` по умолчанию.

#### `func (c Config) Middleware() grpc.Middleware`

Создаёт middleware для трассировки gRPC-запросов. Возвращаемый middleware:

- Извлекает контекст трассировки из входящих метаданных;
- Инициирует span с именем, основанным на методе gRPC-запроса (если доступно);
- Добавляет атрибут `request.id`;
- Фиксирует ошибку в span, если она произошла;
- Завершает span после выполнения запроса.

Если трассировка отключена (noop-провайдер), возвращается no-op middleware.

## Usage

### Default usage flow

```go
package main

import (
    "context"

    "github.com/txix-open/isp-kit/grpc"
    "github.com/txix-open/isp-kit/grpc/endpoint"

    "github.com/txix-open/isp-kit/log"

    "github.com/txix-open/isp-kit/validator"

    "github.com/txix-open/isp-kit/observability/tracing/grpc/server_tracing"
)

func main(){
    ...

    tracingCfg := server_tracing.NewConfig()
    server := grpc.DefaultServer()

    logger, _ := log.New()
    paramMappers := []endpoint.ParamMapper{
		ContextParam(),
		AuthDataParam(),
	}
	middlewares := append(
		[]grpc.Middleware{
			RequestId(),
			server_tracing.NewConfig().Middleware(),
			ErrorHandler(logger),
			Recovery(),
		},
		restMiddlewares...,
	)
	wrapper := NewWrapper(
		paramMappers,
		JsonRequestExtractor{Validator: validator.Default},
		JsonResponseMapper{},
	).WithMiddlewares(middlewares...)

    mux := grpc.NewMux()
    mux.Handle("some_endpoint", wrapper.Endpoint(func(ctx context.Context) error {
        // реализация
        return nil
    }))

    server.Upgrade(mux)

    server.ListenAndServe(":8080")
}
```

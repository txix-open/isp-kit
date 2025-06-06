# Package `endpoint`

Пакет `endpoint` предоставляет инструменты для создания gRPC-обработчиков с автоматической сериализацией JSON,
валидацией запросов, middleware-цепочками и интеграцией метрик/логирования. Реализует паттерн "Endpoint" для изоляции
бизнес-логики от транспортного слоя.

## Types

### Wrapper

Основная структура для создания цепочек обработки запросов. Объединяет:

- Маппинг параметров
- Middleware (логирование, метрики, восстановление после паник)
- Сериализацию/валидацию JSON

**Methods:**

#### `NewWrapper(paramMappers []ParamMapper, bodyExtractor RequestBodyExtractor, bodyMapper ResponseBodyMapper) Wrapper`

Создает конфигурацию для обработчиков.  
Параметры:

- `paramMappers` – правила извлечения параметров из контекста/метаданных
- `bodyExtractor` – десериализация и валидация тела запроса
- `bodyMapper` – сериализация ответа

#### `(m Wrapper) Endpoint(f any) grpc.HandlerFunc`

Преобразует пользовательскую функцию в gRPC-обработчик.  
Функция `f` может содержать параметры:

- `context.Context`
- `grpc.AuthData` (метаданные аутентификации)
- Пользовательский тип (десериализуется из JSON-тела)

#### `(m Wrapper) WithMiddlewares(middlewares ...grpc.Middleware) Wrapper`

Добавляет middleware в цепочку обработки.

### Caller

Внутренняя структура для вызова пользовательских обработчиков. Автоматически:

- Извлекает параметры
- Вызывает middleware
- Обрабатывает ошибки

**Methods:**

####

`NewCaller(f any,bodyExtractor RequestBodyExtractor, bodyMapper ResponseBodyMapper, paramMappers map[string]ParamMapper) (*Caller, error)`

Создает экземпляр `Caller` для функции-обработчика.

#### `(h *Caller) Handle(ctx context.Context, message *isp.Message) (*isp.Message, error)`

Вызывает функцию-обработчик, передавая ей аргументы, извлеченные из запроса.

### JsonRequestExtractor

Реализация интерфейса `RequestBodyExtractor` для работы с JSON. Валидирует данные с помощью объекта, реализующего
интерфейс `Validator`.

**Methods:**

####

`(j JsonRequestExtractor) Extract(ctx context.Context, message *isp.Message, reqBodyType reflect.Type) (reflect.Value, error)`

Десериализует и валидирует JSON-тело запроса через `Validator`.

### JsonResponseMapper

Сериализует ответ обработчика в JSON-тело gRPC-сообщения.

**Methods:**

#### `(j JsonResponseMapper) Map(result any) (*isp.Message, error)`

Конвертирует результат обработчика в gRPC-сообщение в формате JSON.

## Functions

#### `DefaultWrapper(logger log.Logger, restMiddlewares ...grpc.Middleware) Wrapper`

Создает предварительно настроенную обертку (`Wrapper`) для gRPC-обработчиков с базовыми middleware и параметрами.
Упрощает создание эндпоинтов, включая валидацию, логирование, метрики и обработку ошибок "из коробки".

Стандартные middleware:

- `RequestId` – добавляет в контекст requestId, который берет из заголовка x-request-id. Генерирует
  новый, если не находит.
- `Metrics` – собирает метрики: время выполнения, статусы, размеры тел.
- `Tracing` – интеграция с трейсингом (OpenTelemetry).
- `ErrorHandler` – перехватывает и обрабатывает ошибки. Ошибки типа `GrpcError` возвращают структурированный ответ.
  Остальные ошибки логируются и возвращаются как Internal Server Error с gRPC-кодом 13.
- `Recovery` – предотвращает падение сервера при панике в обработчике, преобразуя ее в ошибку.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/endpoint"
	log2 "github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/shutdown"
)

type getUserRequest struct {
	Id string
}

type user struct {
	Id   string
	Name string
}

func getUser(ctx context.Context, authData grpc.AuthData, req getUserRequest) (*user, error) {
	appId, _ := authData.ApplicationId()
	log.Printf("Request from app with id %d", appId)

	/* put here business logic */

	return &user{Id: req.Id, Name: "Alice"}, nil
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	mux := grpc.NewMux()
	wrapper := endpoint.DefaultWrapper(logger)
	mux.Handle("/get_user", wrapper.Endpoint(getUser))

	srv := grpc.DefaultServer()
	srv.Upgrade(mux)

	shutdown.On(func() { /* waiting for SIGINT & SIGTERM signals */
		log.Println("shutting down...")
		srv.Shutdown()
		log.Println("shutdown completed")
	})

	err = srv.ListenAndServe(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

```
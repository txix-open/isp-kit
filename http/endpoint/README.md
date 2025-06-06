# Package `endpoint`

Пакет `endpoint` предоставляет инструменты для создания HTTP-обработчиков с поддержкой middleware, валидации данных,
логирования и метрик. Упрощает разработку API, автоматизируя рутинные задачи.

## Types

### Wrapper

Структура для создания оберток вокруг обработчиков. Объединяет:

- Параметры функции (контекст, запрос, тело и т.д.).
- Middleware (логирование, метрики, трейсинг).
- Валидацию и сериализацию данных.

**Methods:**

####

`NewWrapper(paramMappers []ParamMapper, bodyExtractor RequestBodyExtractor, bodyMapper ResponseBodyMapper, logger log.Logger) Wrapper`

Конструктор обертки с указанными параметрами.

#### `(m Wrapper) Endpoint(f any) http.HandlerFunc`

Преобразовать функцию-обработчик в HTTP-обработчик.

#### `(m Wrapper) WithMiddlewares(middlewares ...http2.Middleware) Wrapper`

Билдер-метод для добавления middleware в обертку.

### JsonRequestExtractor

Извлекает JSON из тела запроса и валидирует его с помощью объекта, реализующего интерфейс `Validator`.

**Methods:**

####

`(j JsonRequestExtractor) Extract(ctx context.Context, reader io.Reader, reqBodyType reflect.Type) (reflect.Value, error)`

Десериализовать JSON и проверить данные через `Validator`.

### JsonResponseMapper

Сериализует ответ в JSON и добавляет заголовки.

**Methods:**

#### `(j JsonResponseMapper) Map(ctx context.Context, result any, w http.ResponseWriter) error`

Упаковка результата в JSON-ответ.

### Caller

Структура `Caller` предоставляет механизм для вызова функций-обработчиков HTTP-запросов, автоматически преобразуя
параметры запроса в аргументы функции. Интегрируется с валидацией, извлечением данных и middleware.

**Methods:**

####

`NewCaller(f any, bodyExtractor RequestBodyExtractor, bodyMapper ResponseBodyMapper, paramMappers map[string]ParamMapper) (*Caller, error)`

Создает экземпляр `Caller` для функции-обработчика.

#### `(h *Caller) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) error`

Вызывает функцию-обработчик, передавая ей аргументы, извлеченные из запроса.

## Functions

#### `DefaultWrapper(logger log.Logger, logMiddleware LogMiddleware, restMiddlewares ...http.Middleware) Wrapper`

Создает предварительно настроенную обертку (`Wrapper`) для HTTP-обработчиков с базовыми middleware и параметрами.
Упрощает создание эндпоинтов, включая валидацию, логирование, метрики и обработку ошибок "из коробки".

Стандартные Middleware:

- `MaxRequestBodySize` – ограничивает размер тела запроса (по умолчанию 64 МБ).
- `RequestId` – добавляет в контекст requestId, который берет из заголовка x-request-id. Генерирует
  новый, если не находит.
- `LogMiddleware` – логирует данные запросов и ответов.
- `Metrics` – собирает метрики: время выполнения, статус-коды,
  размеры тел.
- `Tracing` – интеграция с трейсингом (OpenTelemetry).
- `ErrorHandler` – перехватывает и обрабатывает ошибки. Ошибки типа `HttpError` возвращают структурированный ответ.
  Остальные ошибки логируются и возвращаются как 500 Internal Server Error.
- `Recovery` – предотвращает падение сервера при панике в обработчике, преобразуя ее в ошибку.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/http"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/http/router"
	log2 "github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/shutdown"
)

type userRequest struct {
	Name string
}

type userResponse struct {
	Id int
}

func foo(ctx context.Context, req userRequest) (userResponse, error) {
	/* put here some business logic */
	return userResponse{Id: 88}, nil
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	srv := http.NewServer(logger)

	wrapper := endpoint.DefaultWrapper(logger, httplog.Log(logger, true))
	r := router.New()
	r.POST("/foo", wrapper.Endpoint(foo))

	srv.Upgrade(r)

	shutdown.On(func() { /* waiting for SIGINT & SIGTERM signals */
		log.Println("shutting down...")
		_ = srv.Shutdown(context.Background())
		log.Println("shutdown completed")
	})

	err = srv.ListenAndServe(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

```
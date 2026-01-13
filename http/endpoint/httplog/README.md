# Package `httplog`

Пакет `httplog` предоставляет middleware для логирования HTTP-запросов и ответов. Поддерживает гибкую настройку: выбор
типов контента для логирования тела, управление детализацией и интеграцию с буферизацией данных.

## Functions

#### `Log(logger log.Logger, logBody bool) endpoint.LogMiddleware`

Создает middleware с базовыми настройками:

- Логирует тело запроса/ответа, если `logBody = true`.
- По умолчанию логирует только для `application/json` и `text/xml`.

#### `CombinedLog(logger log.Logger, logBody bool) endpoint.LogMiddleware`

Создает middleware с базовыми настройками:

- Логирует тело запроса/ответа, если `logBody = true`.
- По умолчанию логирует только для `application/json` и `text/xml`.
- Собирает запрос и ответ в 1 лог.

#### `LogWithOptions(logger log.Logger, opts ...Option) endpoint.LogMiddleware`

Создает middleware с кастомными настройками через опции:

- `WithContentTypes(logBodyContentTypes []string) Option` – Задает типы контента, для которых логируется тело (например,
  `application/json`).
- `WithLogBody(logBody bool) Option` – Включает/отключает логирование тела запроса и ответа.
- `WithLogResponseBody(logResponseBody bool) Option` – Включает/отключает логирование тела ответа.
- `WithLogRequestBody(logRequestBody bool) Option` – Включает/отключает логирование тела запроса.
- `WithCombinedLog(enable bool) Option` – Включает/отключает сборку запроса/ответа в 1 лог.

## Usage

### Default log middleware

```go
package main

import (
	"log"

	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	log2 "github.com/txix-open/isp-kit/log"
)

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	/* create wrapper with default logging middleware */
	wrapper := endpoint.DefaultWrapper(logger, httplog.Log(logger, true))
}

```

### Customize log middleware

```go
package main

import (
	"log"

	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	log2 "github.com/txix-open/isp-kit/log"
)

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	/* create wrapper with custom logging middleware */
	wrapper := endpoint.DefaultWrapper(logger, httplog.LogWithOptions(
		logger,
		httplog.WithLogRequestBody(true),                         // enable logging request body
		httplog.WithLogResponseBody(false),                       // disable logging response body
		httplog.WithContentTypes([]string{"application/custom"}), // only for `application/custom` content-type
	))
}

```
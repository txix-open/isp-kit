# Package `grpclog`

Пакет `grpclog` предоставляет middleware для логирования gRPC-запросов и ответов с гибкой настройкой детализации.
Поддерживает логирование тел сообщений и времени выполнения.

## Functions

#### `Log(logger log.Logger, logBody bool) grpc.Middleware`
Создает middleware с базовыми настройками. Логирует тела запроса и ответа, если `logBody = true`.

#### `CombinedLog(logger log.Logger, logBody bool) grpc.Middleware`
Создает middleware с базовыми настройками. Логирует тела запроса и ответа, если `logBody = true`, собирает запрос и ответ в 1 лог.

#### `LogWithOptions(logger log.Logger, opts ...Option) grpc.Middleware`
Создает middleware с кастомными настройками через опции:

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

	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/grpc/endpoint/grpclog"
	log2 "github.com/txix-open/isp-kit/log"
)

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	/* create wrapper with default logging middleware */
	wrapper := endpoint.DefaultWrapper(logger, grpclog.Log(logger, true))
}

```

### Customize log middleware

```go
package main

import (
	"log"

	"github.com/txix-open/isp-kit/grpc/endpoint"
	"github.com/txix-open/isp-kit/grpc/endpoint/grpclog"
	log2 "github.com/txix-open/isp-kit/log"
)

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	/* create wrapper with custom logging middleware */
	wrapper := endpoint.DefaultWrapper(logger, grpclog.LogWithOptions(
		logger,
		grpclog.WithLogRequestBody(true),   // enable logging request body
		grpclog.WithLogResponseBody(false), // disable logging response body
		grpclog.WithLogBody(true),          // enable logging request's & response's bodies
	))
}

```
# Package `pprof`

Пакет `pprof` для интеграции Go-профилировщика pprof с HTTP-сервером.

## Functions

#### `RegisterHandlers(prefix string, muxer Muxer)`

Зарегистрировать обработчики pprof с указанным префиксом URL для переданного мультиплексора, реализующего интерфейс `Muxer`.

#### `Endpoints(prefix string) []string`

Получить список всех зарегистрированных эндпоинтов.

## Usage

### Default usage flow

```go
package main

import (
	"log"

	"github.com/txix-open/isp-kit/infra"
	"github.com/txix-open/isp-kit/infra/pprof"
	"github.com/txix-open/isp-kit/shutdown"
)

func main() {
	srv := infra.NewServer()
	pprof.RegisterHandlers("/debug", srv)

	shutdown.On(func() { /* waiting for SIGINT & SIGTERM signals */
		log.Println("shutting down...")
		srv.Shutdown()
		log.Println("shutdown completed")
	})

	err := srv.ListenAndServe(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

```
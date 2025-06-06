# Package `infra`

Пакет `infra` предоставляет обертку над http-сервером из стандартной библиотеки `net/http`.

## Types

### Server

Структура-обертка http-сервера.

**Methods:**

#### `NewServer() *Server`

Создать экземпляр http-сервера.

#### `(s *Server) Handle(pattern string, handler http.Handler)`

Зарегистрировать обработчик.

#### `(s *Server) HandleFunc(pattern string, handler http.HandlerFunc)`

Зарегистрировать функцию-обработчик.

#### `(s *Server) ListenAndServe(address string) error`

Запустить http-сервер. Обрабатывает `http.ErrServerClosed` как нормальное завершение.

#### `(s *Server) Shutdown() error`

Остановить работу http-сервера.

## Usage

### Default usage flow

```go
package main

import (
	"log"
	"net/http"

	"github.com/txix-open/isp-kit/infra"
	"github.com/txix-open/isp-kit/shutdown"
)

func main() {
	srv := infra.NewServer()
	srv.HandleFunc("/foo", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

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
# Package `http`

Пакет `http` предоставляет инструменты для создания и управления HTTP-сервером с поддержкой middleware, обработки
запросов и гибкой настройки. Интегрируется с системой логирования и обработки ошибок.

## Types

### Server

Структура Server управляет жизненным циклом HTTP-сервера, включая запуск, остановку и динамическое обновление
обработчиков.

**Methods:**

#### `NewServer(logger log.Logger, opts ...ServerOption) *Server`

Конструктор сервера. Дополнительно принимает опции:

- `WithServer(server *http.Server) ServerOption` – использование объекта предварительно настроенного HTTP-сервера из
  стандартной библиотеки `net/http`.

#### `(s *Server) Upgrade(handler http.Handler)`

Атомарно заменить текущий обработчик запросов на новый.

#### `(s *Server) ListenAndServe(address string) error`

Запустить сервер на указанном адресе.

#### `(s *Server) Serve(listener net.Listener) error`

Запустить сервер на существующем listener.

#### `(s *Server) Shutdown(ctx context.Context) error`

Остановить сервер, завершив все активные соединения.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"
	"net/http"

	http2 "github.com/txix-open/isp-kit/http"
	log2 "github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/shutdown"
)

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	srv := http2.NewServer(logger)
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("hello world!!"))
	})

	srv.Upgrade(mux)

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
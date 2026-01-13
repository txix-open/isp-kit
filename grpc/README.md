# Package `grpc`

Пакет `grpc` предоставляет инструменты для создания и управления gRPC-серверами, включая обработку аутентификации,
маршрутизацию запросов и интеграцию с метаданными.
Поддерживает гибкую настройку обработчиков, безопасное обновление сервисов и работу с заголовками.

## Types

### Server

gRPC-сервер с поддержкой динамического обновления обработчиков.

**Methods:**

#### `DefaultServer(restOptions ...grpc.ServerOption) *Server`

Создать gRPC-сервер с настройками по умолчанию (максимальный размер сообщения 64 МБ).

#### `NewServer(opts ...grpc.ServerOption) *Server`

Создать кастомный сервер с указанными опциями gRPC.

#### `(s *Server) Upgrade(service isp.BackendServiceServer)`

Атомарно обновить обработчик сервиса без остановки сервера.

#### `(s *Server) ListenAndServe(address string) error`

Запустить сервер на указанном адресе.

#### `(s *Server) Serve(listener net.Listener) error`

Запустить сервер на существующем listener.

#### `(s *Server) Shutdown()`

Остановить сервер

### Mux

Мультиплексор для маршрутизации gRPC-запросов. Регистрирует обработчики по имени эндпоинта.

**Methods:**

#### `NewMux() *Mux`

Создать новый мультиплексор.

#### `(m *Mux) Handle(endpoint string, handler HandlerFunc) *Mux`

Зарегистрировать обработчик для указанного эндпоинта.

#### `(m *Mux) Request(ctx context.Context, message *isp.Message) (*isp.Message, error)`

Обрабатывает входящий запрос, определяя эндпоинт через заголовок `ProxyMethodNameHeader`.

### AuthData

Структура для работы с метаданными аутентификации в gRPC-запросах.
Оборачивает `metadata.MD` и предоставляет методы для извлечения идентификаторов из заголовков.

**Methods:**

#### `(i AuthData) SystemId() (int, error)`

Получить идентификатор системы из заголовка `x-system-identity`.

#### `(i AuthData) DomainId() (int, error)`

Получить идентификатор домена из заголовка `x-domain-identity`.

#### `(i AuthData) ServiceId() (int, error)`

Получить идентификатор сервиса из заголовка `x-service-identity`.

#### `(i AuthData) ApplicationId() (int, error)`

Получить идентификатор системы из заголовка `x-application-identity`.

#### `(i AuthData) ApplicationName() (int, error)`

Получить название системы из заголовка `x-application-name`.

#### `(i AuthData) UserId() (int, error)`

Получить идентификатор пользователя из заголовка `x-user-identity`.

#### `(i AuthData) DeviceId() (int, error)`

Получить идентификатор устройства из заголовка `x-device-identity`.

#### `(i AuthData) UserToken() (string, error)`

Получить токен пользователя из заголовка `x-user-token`.

#### `(i AuthData) DeviceToken() (string, error)`

Получить токен устройства из заголовка `x-device-token`.

## Functions

#### `StringFromMd(key string, md metadata.MD) (string, error)`

Извлекает строковое значение из метаданных по ключу.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/grpc"
	"github.com/txix-open/isp-kit/grpc/isp"
	"github.com/txix-open/isp-kit/shutdown"
)

func main() {
	mux := grpc.NewMux()
	mux.Handle("/get_users", func(ctx context.Context, msg *isp.Message) (*isp.Message, error) {
		/* put here business logic */
		return new(isp.Message), nil
	})

	srv := grpc.DefaultServer()
	srv.Upgrade(mux)

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
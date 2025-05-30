# Package `httpt`

Пакет `httpt` предназначен для упрощения написания HTTP-тестов (юнит- и интеграционных) с использованием клиента и сервера на базе `isp-kit`.

## Types

### MockServer

Структура `MockServer` представляет собой обёртку над HTTP-сервером, позволяющую удобно мокать отдельные endpoint'ы с использованием `endpoint.Wrapper` и роутера `router.Router`.

**Methods:**

#### `NewMock(t *test.Test) *MockServer`

Создаёт `MockServer`, автоматически запускает сервер и регистрирует его закрытие в `t.Cleanup`. Подходит для юнит-тестов с ручной маршрутизацией и моками.

#### `(m *MockServer) POST(path string, handler any) *MockServer`

Регистрирует POST-обработчик на указанный путь.

#### `(m *MockServer) GET(path string, handler any) *MockServer`

Регистрирует GET-обработчик на указанный путь.

#### `(m *MockServer) Mock(method string, path string, handler any) *MockServer`

Регистрирует обработчик на указанный путь и метод с использованием `Wrapper.Endpoint`.

#### `(m *MockServer) Client(opts ...httpcli.Option) *httpcli.Client`

Создаёт HTTP-клиент, настроенный на тестовый сервер.

#### `(m *MockServer) BaseURL() string`

Возвращает базовый URL тестового сервера.

## Functions

### `TestServer(t *test.Test, handler http.Handler, opts ...httpcli.Option) (*httptest.Server, *httpcli.Client)`

Создаёт `httptest.Server` с указанным HTTP-обработчиком и возвращает его вместе с клиентом `httpcli.Client`. Подходит для end-to-end тестов.

Автоматически управляет закрытием ресурсов (`srv.Close`) по окончании теста.

## Usage

### Example usage in test

```go
package mypkg_test

import (
	"net/http"
	"testing"

	"github.com/txix-open/isp-kit/http/httpt"
	"github.com/txix-open/isp-kit/test"
)

func TestExample(t *testing.T) {
	testCtx := test.New(t)
	srv, cli := httpt.TestServer(testCtx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	resp, err := cli.Get("/")
	testCtx.Assert().NoError(err)
	testCtx.Assert().Equal(200, resp.StatusCode)
	testCtx.Assert().BodyEqual(resp, "ok")
	_ = srv // для дополнительных проверок, если нужно
}
```

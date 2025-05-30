# Package `grpct`

Пакет `grpct` предназначен для упрощения написания юнит- и интеграционных тестов gRPC-сервисов на базе `isp-kit`.

## Types

### MockServer

Структура `MockServer` представляет собой обёртку над gRPC-сервером, позволяющую удобно мокать отдельные endpoint'ы.

**Methods:**

#### `NewMock(t *test.Test) (*MockServer, *client.Client)`

Создаёт `MockServer` и соответствующий gRPC-клиент. Полезно для юнит-тестов и простых сценариев мокирования.

#### `(m *MockServer) Mock(endpoint string, handler any) *MockServer`

Регистрирует указанный `handler` на переданный `endpoint` с использованием `endpoint.Wrapper`.

## Functions

### `TestServer(t *test.Test, service isp.BackendServiceServer) (*grpc.Server, *client.Client)`

Создаёт полноценный `grpc.Server` с переданным `BackendServiceServer` и возвращает его вместе с клиентом. Подходит для end-to-end тестов.

Автоматически управляет закрытием ресурсов (`Shutdown` сервера и `Close` клиента) по окончании теста.

## Usage

### Create mock gRPC server and mock handler

```go
package mypkg_test

import (
	"testing"

	"github.com/txix-open/isp-kit/grpct"
	"github.com/txix-open/isp-kit/test"
)

func TestExample(t *testing.T) {
	testCtx := test.New(t)
	mockSrv, cli := grpct.NewMock(testCtx)
	mockSrv.Mock("/example.Service/Method", handlerFn)

	resp, err := cli.Invoke(testCtx.Context(), "/example.Service/Method", req, &reply)
	testCtx.Assert().NoError(err)
	testCtx.Assert().Equal(expected, reply)
}
```

### Create test gRPC server from implementation

```go
import (
	"testing"

	"github.com/txix-open/isp-kit/grpct"
	"github.com/txix-open/isp-kit/test"
)

func TestRealService(t *testing.T) {
	testCtx := test.New(t)
	srv, cli := grpct.TestServer(testCtx, myServiceImpl)

	resp, err := cli.Invoke(testCtx.Context(), "/my.Service/Endpoint", req, &reply)
	testCtx.Assert().NoError(err)
	testCtx.Assert().Equal(expected, reply)

	_ = srv // используется для дополнительных проверок, если необходимо
}
```

# Package `client`

Пакет `client` предоставляет гибкий gRPC-клиент с поддержкой балансировки нагрузки, middleware для логирования, метрик,
трассировки и динамическим обновлением списка хостов.

## Types

### Client

Основная структура клиента для взаимодействия с gRPC-серверами. Поддерживает балансировку нагрузки, цепочки middleware и
динамическое обновление списка хостов.

**Methods:**

#### `New(initialHosts []string, opts ...Option) (*Client, error)`

Создать новый клиент с указанными начальными хостами и опциями:

- `WithMiddlewares(middlewares ...request.Middleware) Option` – добавить middleware в цепочку обработки запроса.
- `WithDialOptions(dialOptions ...grpc.DialOption) Option` – опция для передачи параметров подключения gRPC (например,
  TLS, таймауты)

#### `(cli *Client) Invoke(endpoint string) *request.Builder`

Инициирует запрос к указанному эндпоинту. Возвращает билдер для выполнения запроса.

#### `(cli *Client) Upgrade(hosts []string)`

Атомарно обновить список хостов.

#### `(cli *Client) Close() error`

Закрыть соединение с сервером.

#### `(cli *Client) BackendClient() isp.BackendServiceClient`

Получить низкоуровневый клиент для прямого взаимодействия с gRPC-сервисом.

## Functions

#### `Default(restMiddlewares ...request.Middleware) (*Client, error)`

Создать клиент с настройками по умолчанию:

- Максимальный размер сообщения 64 МБ.
- Middleware для генерации requestId, сбора метрик через `grpc_metrics.ClientStorage` и трейсинга

#### `RequestId() request.Middleware`

Middleware для автоматической генерации requestId для передачи в заголовках. Если в контексте будет указан requestId, то
передаваться будет именно он.

#### `Log(logger log.Logger, logBody bool) request.Middleware`

Middleware для логирования запросов и ответов. Логирует тело запроса/ответа, если `logBody = true`.

#### `Metrics(storage MetricStorage) request.Middleware`

Middleware для сбора метрик длительности запросов.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/grpc/client"
	log2 "github.com/txix-open/isp-kit/log"
)

type user struct {
	Id   string
	Name string
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	cli, err := client.Default(client.Log(logger, true))
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()

	cli.Upgrade([]string{"host1:8080", "host2:9000"})

	users := make([]user, 0)
	err = cli.Invoke("/get_users").
		JsonResponseBody(&users).
		Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, u := range users {
		/* handle each user */
	}
}

```
# Package `request`

Пакет `request` предоставляет билдер для построения gRPC-запросов с поддержкой JSON-сериализации, таймаутов, метаданных
и цепочек обработки через middleware. Интегрируется с клиентом из пакета [`client`](../client.go).

## Types

### Builder

Структура для пошагового конструирования gRPC-запроса. Обрабатывает сериализацию/десериализацию JSON, управление
контекстом и метаданными.

**Methods:**

#### `NewBuilder(roundTripper RoundTripper, endpoint string) *Builder`

Создать новый билдер для указанного эндпоинта.

#### `(req *Builder) ApplicationId(appId int) *Builder`

Установить идентификатор системы в заголовок `x-application-identity`.

#### `(req *Builder) JsonRequestBody(reqBody any) *Builder`

Задать тело запроса в формате JSON. Объект будет сериализован автоматически при помощи функций из пакета [
`json`](../../json) (маршалинг без тэгов в camelCase).

#### `(req *Builder) JsonResponseBody(respPtr any) *Builder`

Задать указатель для десериализации тела ответа из JSON.

#### `(req *Builder) Timeout(timeout time.Duration) *Builder`

Установить тайм-аут для запроса. По умолчанию: 15 секунд.

#### `(req *Builder) AppendMetadata(k string, v ...string) *Builder`

Добавить кастомные метаданные в запрос.

#### `(req *Builder) Do(ctx context.Context) error`

Выполнить запрос. Автоматически:

1. Сериализует тело запроса в JSON
2. Добавляет системные заголовки (`x-application-identity`, `proxy_method_name`)
3. Обрабатывает цепочку middleware
4. Десериализует ответ (если задан `responsePtr`)

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
	"time"

	"github.com/txix-open/isp-kit/grpc/client"
	log2 "github.com/txix-open/isp-kit/log"
)

type getUserRequest struct {
	Id string
}

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

	u := new(user)
	err = cli.Invoke("/get_user").
		ApplicationId(88).
		JsonRequestBody(getUserRequest{Id: "some-user-id"}).
		JsonResponseBody(u).
		Timeout(5 * time.Second).
		Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	
	log.Printf("username: %s\n", u.Name)
}

```
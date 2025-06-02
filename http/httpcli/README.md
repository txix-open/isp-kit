# Package `httpcli`

Пакет `httpcli` предоставляет расширенный HTTP-клиент с поддержкой middleware, ретраев, различных типов запросов и
обработки ошибок.
Интегрируется с системами логирования, метриками и политиками повторов.

## Types

### Client

Основная структура для выполнения HTTP-запросов. Управляет настройками транспорта, middleware и глобальными параметрами.

**Methods:**

#### `New(opts ...Option) *Client`

Конструктор клиента с настройками по умолчанию. Принимает опции:

- `WithMiddlewares(mws ...Middleware) Option` – добавляет цепочку middleware к запросам клиента.

#### `NewWithClient(cli *http.Client, opts ...Option) *Client`

Конструктор клиента на основе базового клиента из стандартной библиотеки `net/http`.

#### `(c *Client) GlobalRequestConfig() *GlobalRequestConfig`

Получить глобальную конфигурацию запросов (таймауты, базовый URL, заголовки и т.д.).

#### `(c *Client) Post(url string) *RequestBuilder`

Создать билдер для POST-запроса. Аналогичные методы: `Get()`, `Put()`, `Delete()`, `Patch()`.

#### `(c *Client) Execute(ctx context.Context, builder *RequestBuilder) (*Response, error)`

Выполнить HTTP-запрос из переданного билдера.

### RequestBuilder

Структура для построения HTTP-запросов с цепочкой методов.

**Methods:**

#### `(b *RequestBuilder) Header(name string, value string) *RequestBuilder`

Добавить заголовок к запросу.

#### `(b *RequestBuilder) BaseUrl(baseUrl string) *RequestBuilder`

Установить базовый url в запросе на переданный.

#### `(b *RequestBuilder) Url(url string) *RequestBuilder`

Установить url запроса на переданный.

#### `(b *RequestBuilder) Method(method string) *RequestBuilder`

Изменить HTTP-метод запроса на указанный.

#### `(b *RequestBuilder) Cookie(cookie *http.Cookie) *RequestBuilder`

Добавить cookie к запросу.

#### `(b *RequestBuilder) RequestBody(body []byte) *RequestBuilder`

Добавить тело запроса.

#### `(b *RequestBuilder) JsonRequestBody(value any) *RequestBuilder`

Добавить тело запроса в формате JSON. Использует пакет [`json`](../../json), поэтому может маршалить без тэгов в
camelCase.

#### `(b *RequestBuilder) JsonResponseBody(responsePtr any) *RequestBuilder`

Записать JSON-ответ на запрос по переданному указателю. Операция выполнится только при статус кодах от 200 до 299.

#### `(b *RequestBuilder) FormDataRequestBody(data map[string][]string) *RequestBuilder`

Добавить тело запроса в формате формы.

#### `(b *RequestBuilder) BasicAuth(ba BasicAuth) *RequestBuilder`

Добавить базую аутентификацию в запрос.

#### `(b *RequestBuilder) QueryParams(queryParams map[string]any) *RequestBuilder`

Добавить query-параметры в запрос.

#### `(b *RequestBuilder) Retry(cond RetryCondition, retryer Retryer) *RequestBuilder`

Задать политику ретраев по заданному условию. Пример условия: `httpcli.IfErrorOr5XXStatus()`

#### `(b *RequestBuilder) MultipartRequestBody(data *MultipartData) *RequestBuilder`

Устанавить multipart/form-data тело запроса (для передачи файлов). Не поддерживает `Request.Body` в middleware и
игнорирует ретраи.

#### `(b *RequestBuilder) StatusCodeToError() *RequestBuilder`

Возвращает `ErrorResponse` как ошибку при `Response.IsSuccess = true`.

#### `(b *RequestBuilder) Timeout(timeout time.Duration) *RequestBuilder`

Установить тайм-аут для каждой попытки запроса, тайм-аут по умолчанию 15 секунд.

#### `(b *RequestBuilder) Middlewares(middlewares ...Middleware) *RequestBuilder`

Добавить middleware в цепочку запроса.

#### `(b *RequestBuilder) Do(ctx context.Context) (*Response, error)`

Выполнить запрос собранный билдером.

#### `(b *RequestBuilder) DoWithoutResponse(ctx context.Context) error`

То же самое, что и `Do`, но записывает ответ по переданному раннее в билдере указателю.

#### `(b *RequestBuilder) DoWithoutResponse(ctx context.Context) error`

То же самое, что и `Do`, но не возвращает объект `Response`. Записывает тело ответа по переданному в метод
`JsonResponseBody` указателю.

#### `(b *RequestBuilder) DoAndReadBody(ctx context.Context) ([]byte, int, error)`

То же самое, что и `Do`, но считывает из ответа только тело и статус код, а затем возвращает их.

## Usage

### Default

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/http/httpcli"
)

type user struct {
	Id   string
	Name string
}

func main() {
	cli := httpcli.New()
	u := new(user)

	resp, err := cli.Get("https://api.example.com/users/1").
		JsonResponseBody(u).
		Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Close()

	if !resp.IsSuccess() {
		log.Fatalf("invalid status code: %d", resp.StatusCode())
	}
	log.Println(u.Name) /* result decoded from JSON */
}

```

### Without response

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/http/httpcli"
)

type user struct {
	Id   string
	Name string
}

func main() {
	cli := httpcli.New()
	u := new(user)

	err := cli.Get("https://api.example.com/users/1").
		JsonResponseBody(u).
		StatusCodeToError().
		DoWithoutResponse(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println(u.Name) /* result decoded from JSON */
}

```

### Exponential retries

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/retry"
)

type user struct {
	Id   string
	Name string
}

const (
	maxRetryElapsedTime = 5 * time.Second
)

func main() {
	cli := httpcli.New()
	u := user{
		Id:   "x.x",
		Name: "Mark",
	}

	err := cli.Post("https://api.example.com/users").
		JsonRequestBody(u).
		Retry(httpcli.IfErrorOr5XXStatus(), retry.NewExponentialBackoff(maxRetryElapsedTime)).
		StatusCodeToError().
		DoWithoutResponse(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

```
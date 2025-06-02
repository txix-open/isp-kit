# Package `httpclix`

Пакет `httpclix` предоставляет расширенный HTTP-клиент с балансировкой нагрузки, метриками, трейсингом и логированием.
Интегрируется с системами мониторинга и трассировки, поддерживает гибкую настройку middleware.

## Types

### ClientBalancer

Структура для клиента со встроенным балансировщиком HTTP-запросов между несколькими хостами. Управляет подключениями,
обновлением списка хостов и автоматическим добавлением схемы (HTTP/HTTPS).
Расширяет [`httpcli.Client`](../httpcli/client.go) и использует пакет [`lb`](../../lb) для балансировки.

**Methods:**

Содержит все методы клиента-предшественника.

#### `NewClientBalancer(initialHosts []string, opts ...Option) *ClientBalancer`

Конструктор HTTP-клиента с балансировщиком. Принимает начальный список хостов и опции для настройки:

- `WithClientOptions(opts ...httpcli.Option) Option` – добавить опции для родительского клиента из пакета [
  `httpcli`](../httpcli).
- `WithHttpsSchema() Option` – автоматически добавлять к хостам https схему, вместо http.
- `WithClient(cli *http.Client) Option` – создание http-клиента на основе базового клиента из стандартной библиотеки
  `net/http`.

#### `(c *ClientBalancer) Upgrade(hosts []string)`

Обновляет список активных хостов. Автоматически добавляет схему, если она не указана.

### ClientTracer

Структура `ClientTracer` предоставляет инструменты для трейсинга HTTP-запросов, включая измерение времени выполнения
ключевых этапов. Используется совместно с httptrace.ClientTrace для интеграции в HTTP-клиент.

**Methods:**

#### `NewClientTracer(clientStorage *metrics.ClientStorage, endpoint string) *ClientTracer`

Создает трейсер для интеграции с httptrace.

#### `(cli *ClientTracer) ClientTrace() *httptrace.ClientTrace`

Создает и возвращает объект httptrace.ClientTrace с хуками для отслеживания этапов запроса:

- `DNSStart` – Засекает время начала DNS-запроса.
- `DNSDone` – Рассчитывает длительность DNS-запроса и сохраняет в метрики через `clientStorage.ObserveDnsLookup`.
- `ConnectStart` – Засекает начало установки TCP-соединения.
- `ConnectDone` – Сохраняет длительность установки соединения через `clientStorage.ObserveConnEstablishment`.
- `WroteHeaders` – Начало записи тела запроса (после отправки заголовков).
- `WroteRequest` – Сохраняет длительность записи тела запроса через `clientStorage.ObserveRequestWriting`.
- `GotFirstResponseByte` – Засекает получение первого байта ответа.

#### `(cli *ClientTracer) ResponseReceived()`

Вызывается при полном получении ответа. Рассчитывает длительность чтения ответа и сохраняет в метрики через
`clientStorage.ObserveResponseReading`.

## Functions

#### `Default(opts ...httpcli.Option) *httpcli.Client`

Создает клиент с предустановленными middleware:

- Добавление RequestId.
- Сбор метрик через http_metrics.
- Трейсинг запросов.

#### `DefaultWithBalancer(initialHosts []string, opts ...Option) *ClientBalancer`

Создает клиент-балансировщик с предустановленными middleware:

- Добавление RequestId.
- Сбор метрик через http_metrics.
- Трейсинг запросов.

#### `RequestId() httpcli.Middleware`

Добавляющая заголовок X-Request-Id к запросам middleware. Если requestId отсутствует в контексте — генерирует
новый.

#### `Metrics(storage *http_metrics.ClientStorage) httpcli.Middleware`

Собирающая метрики middleware:

- Время выполнения запроса.
- Статус-коды ответов.
- Ошибки
- Время DNS, установки соединения, записи тела и чтения ответа.

#### `Log(logger log.Logger) httpcli.Middleware`

Middleware для логирования запросов и ответов, включая заголовки и тело (опционально, можно настроить с помощью
контекста).

####

`LogConfigToContext(ctx context.Context, logRequestBody bool, logResponseBody bool, opts ...LogOption) context.Context`

Данная функция добавляет в контекст настройки логирования для HTTP-запросов. Эти настройки определяют, какие данные (
тело, заголовки, дампы) будут записаны в лог при обработке запроса и ответа.

Доступные опции:

- `LogDump(dumpRequest bool, dumpResponse bool) LogOption` – включение/выключение дампа запроса и ответа.
- `LogHeaders(requestHeaders bool, responseHeaders bool) LogOption` – включение/выключение логирования заголовков
  запроса и ответа.

## Usage

### Default

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/http/httpclix"
)

type user struct {
	Id   string
	Name string
}

func main() {
	/* client with metrics, traces & logging */
	cli := httpclix.Default()
	userList := make([]user, 0)

	err := cli.Get("https://api.example.com/users").
		JsonResponseBody(&userList).
		StatusCodeToError().
		DoWithoutResponse(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

```

### Client with balancer

```go
package main

import (
	"context"

	"github.com/txix-open/isp-kit/http/httpclix"
)

type user struct {
	Id   string
	Name string
}

func main() {
	hosts := []string{"host1:8080", "host2:8080", "http://host2:8080"}
	cli := httpclix.NewClientBalancer(hosts, httpclix.WithHttpsSchema())

	/* Requests are distributed between hosts */
	for i := range 5 {
		err := cli.Get("/data").
			JsonResponseBody(&userList).
			StatusCodeToError().
			DoWithoutResponse(context.Background())
		/* some business logic */
	}

	/* update host list */
	cli.Upgrade([]string{"new-host:8080"})
}

```
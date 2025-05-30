# Package `client_tracing`

Пакет `client_tracing` предоставляет middleware для трассировки исходящих HTTP-запросов с использованием OpenTelemetry.
Он интегрируется с клиентом `httpcli` из `isp-kit` и добавляет в запросы span'ы и trace-информацию, в том числе request id и HTTP-атрибуты.

## Types

### Config

Структура `Config` определяет настройки трассировки для исходящих HTTP-запросов.

**Fields:**

#### `Provider tracing.TracerProvider`

Провайдер трассировки (по умолчанию `tracing.DefaultProvider`).

#### `Propagator tracing.Propagator`

Пропагатор контекста (по умолчанию `tracing.DefaultPropagator`).

#### `EnableHttpTracing bool`

Включает детализацию трейсинга через `net/http/httptrace`.

**Methods:**

#### `func NewConfig() Config`

Создаёт конфигурацию трассировки с `provider` и `propagator` по умолчанию.

#### `(c Config) Middleware() httpcli.Middleware`

Создаёт middleware, оборачивающее `httpcli.RoundTripper` для добавления трейсинга в каждый HTTP-запрос.
Выполняются следующие действия:

- Запуск нового спана с именем по шаблону `HTTP call METHOD URL_PATH`.
- Добавление стандартных HTTP-атрибутов в span (метод, URL, статус-код и т.п.).
- Инжекция trace-контекста в заголовки запроса.
- В случае ошибки или получения ответа — установка соответствующего статуса в span.
- При включённом `EnableHttpTracing` — подключение дополнительных `httptrace` хуков.

## Usage

### Default usage flow

```go
package main

import (
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/observability/tracing/http/client_tracing"
)

func main(){
    ...
    client := httpcli.New(
        httpcli.WithMiddleware(
            client_tracing.NewConfig().Middleware(),
        ),
    )
    ...
}
```

# Package `server_tracing`

Пакет `server_tracing` предоставляет middleware для серверной части HTTP-приложения, обеспечивая автоматическое создание span'ов OpenTelemetry для обработки входящих HTTP-запросов.

## Types

### Config

Структура `Config` определяет настройки трассировки для входящих HTTP-запросов.

**Fields:**

#### `Provider tracing.TracerProvider`

Провайдер трассировки (по умолчанию `tracing.DefaultProvider`).

#### `Propagator tracing.Propagator`

Пропагатор контекста (по умолчанию `tracing.DefaultPropagator`).

**Methods:**

#### `NewConfig() Config`

Создаёт конфигурацию с провайдером и пропагатором по умолчанию:

#### `func (c Config) Middleware() http2.Middleware`

Возвращает middleware, который добавляет в контекст обработки запроса OpenTelemetry span с необходимыми атрибутами. При этом:

- Используется `Propagator` для извлечения контекста трассировки из входящего запроса.
- Определяется имя span'а на основе маршрута или метода + пути.
- В span записываются стандартные HTTP-атрибуты, request ID, а также (если используется `buffer.Buffer`) количество прочитанных и записанных байт.
- Ошибки логируются и, если уровень логирования `Error`, также записываются в span через `RecordError`.

## Usage

### Default usage flow

```go
package main

import (
	"github.com/txix-open/isp-kit/observability/tracing/http/server_tracing"
	"github.com/txix-open/isp-kit/http"
)

func main(){
    ...
    tracingCfg := server_tracing.NewConfig()
    mux := http.NewServeMux()
    handler := http.WrapHandler(tracingCfg.Middleware())(func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
        // обработка запроса
        return nil
    })
    mux.Handle("/example", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        _ = handler(r.Context(), w, r)
    }))
    ...
}
```

# Package `requestid`

Пакет `requestid` предоставляет утилиты для генерации и хранения уникального идентификатора запроса (`request-id`) в контексте.

## Constants

### `Header`

Константа `Header` содержит строку `"x-request-id"`, которая представляет собой стандартное имя HTTP-заголовка, используемого для передачи идентификатора запроса.

### `LogKey`

Константа `LogKey` содержит строку `"requestId"` — ключ, под которым `request-id` может сохраняться в логах.

## Types

Данный пакет не экспортирует пользовательских типов.

## Functions

### `ToContext(ctx context.Context, value string) context.Context`

Сохраняет `request-id` в переданном контексте и возвращает новый контекст с сохранённым значением.

### `FromContext(ctx context.Context) string`

Извлекает `request-id` из контекста. Если значение отсутствует, возвращается пустая строка.

### `Next() string`

Генерирует новый случайный `request-id` длиной 16 байт (в hex-представлении — 32 символа).
Использует криптографически безопасный генератор случайных чисел. В случае ошибки вызывает `panic`.

## Usage

```go
import (
	"net/http"
	"github.com/txix-open/isp-kit/requestid"
)

func handler(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get(requestid.Header)
	if id == "" {
		id = requestid.Next()
	}
	r = r.WithContext(requestid.ToContext(r.Context(), id))
	// теперь можно использовать FromContext(r.Context()) в любом месте
}
```

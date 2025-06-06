# Package `buffer`

Пакет `buffer` предоставляет инструменты для буферизации HTTP-запросов и ответов. Используется в middleware для
логирования, метрик и других задач, требующих доступа к телам запросов и ответов без их модификации.

## Types

### Buffer

Структура для перехвата и буферизации данных HTTP-ответа. Реализует интерфейс `http.ResponseWriter`.

**Methods:**

#### `New() *Buffer`

Конструктор буфера. Инициализирует внутренние буферы для тела запроса и ответа.

#### `(m *Buffer) Reset(w http.ResponseWriter)`

Сбрасывает состояние буфера и привязывает его к новому `http.ResponseWriter`.

#### `(m *Buffer) Write(b []byte) (int, error)`

Перехватывает данные, записываемые в ответ, и сохраняет их в буфер.

#### `(m *Buffer) WriteHeader(statusCode int)`

Фиксирует статус-код ответа.

#### `(m *Buffer) ResponseBody() []byte`

Возвращает тело ответа в виде байтов из буфера.

#### `(m *Buffer) RequestBody() []byte`

Возвращает тело запроса в виде байтов из буфера.

#### `(m *Buffer) ReadRequestBody(r io.Reader) error`

Читает тело запроса и сохраняет его в буфер.

#### `(m *Buffer) StatusCode() int`

Возвращает статус-код ответа. Если статус-код не был еще записан, вернется 200 OK.

## Functions

#### `Acquire(w http.ResponseWriter) *Buffer`

Берет буфер из пула и инициализирует его для работы с указанным `http.ResponseWriter`.

#### `Release(w *Buffer)`

Возвращает буфер в пул для повторного использования.

## Usage

### Default usage flow

```go
package main

import (
	"log"
	"net/http"

	"github.com/txix-open/isp-kit/http/endpoint/buffer"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		/* get buffer from pool */
		buf := buffer.Acquire(w)
		defer buffer.Release(buf)

		/* read request body */
		_ = buf.ReadRequestBody(r.Body)
		r.Body = buffer.NewRequestBody(buf.RequestBody())

		next.ServeHTTP(buf, r)

		/* log response body */
		responseBody := buf.ResponseBody()
		log.Printf("Response: %s\n", responseBody)
	})
}

```
# Package `router`

Пакет `router` предоставляет обертку над [`httprouter`](https://github.com/julienschmidt/httprouter) с поддержкой метрик
и удобной регистрацией обработчиков.
Интегрируется с системой мониторинга для отслеживания эндпоинтов.

## Types

### Router

Структура для управления HTTP-маршрутами. Обеспечивает регистрацию обработчиков, сбор метрик и работу с параметрами URL.

**Methods:**

#### `New() *Router`

Конструктор роутера.

#### `(r *Router) GET(path string, handler http.Handler) *Router`

Зарегистрировать обработчик для GET-запросов. Аналогичные методы: `POST()`, `PUT()`, `DELETE()`.

#### `(r *Router) Handler(method string, path string, handler http.Handler) *Router`

Зарегистрировать обработчик для произвольного HTTP-метода.

#### `(r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request)`

Обработать HTTP-запрос.

#### `(r *Router) InternalRouter() *httprouter.Router`

Возвращает внутренний экземпляр `httprouter.Router` для кастомных настроек.

## Functions

#### `ParamsFromRequest(http *http.Request) Params`

Извлечь URL-параметры из запроса.

#### `ParamsFromContext(ctx context.Context) Params`

Извлечь URL-параметры из контекста.

## Usage

### Default usage flow

```go
package main

import (
	"log"
	"net/http"

	"github.com/txix-open/isp-kit/http/router"
)

func main() {
	r := router.New()
	r.GET("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		params := router.ParamsFromRequest(r)
		userId := params.ByName("id") /* get id param */
		log.Println(userId)
		w.WriteHeader(http.StatusOK)
	}))
	_ = http.ListenAndServe(":8080", r)
}

```
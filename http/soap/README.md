# Package `soap`

Пакет `soap` предоставляет инструменты для работы с SOAP-сервисами: обработка запросов, валидация данных, генерация
ошибок в формате SOAP Fault, а также интеграция с метриками и логированием.

## Types

### ActionMux

Мультиплексор для маршрутизации SOAP-запросов на основе заголовка SOAPAction.

**Methods:**

#### `NewActionMux() *ActionMux`

Создает новый мультиплексор.

#### `(m *ActionMux) Handle(actionUri string, handler http.Handler) *ActionMux`

Регистрирует обработчик для указанного SOAPAction.

#### `(m *ActionMux) ServeHTTP(writer http.ResponseWriter, request *http.Request)`

Обрабатывает запрос, определяя действие через заголовок SOAPAction, и вызывает соответствующий обработчик. При
отсутствии действия или неизвестном SOAPAction возвращает SOAP Fault.

### RequestExtractor

Извлекает данные из SOAP-конверта и валидирует их с помощью объекта, реализующего интерфейс `Validator`.

**Methods:**

####

`(j RequestExtractor) Extract(_ context.Context, reader io.Reader, reqBodyType reflect.Type) (reflect.Value, error)`

Десериализация XML и проверка данных через `Validator`.

### ResponseMapper

Формирует SOAP-ответ из данных, добавляя заголовки и сериализуя в XML.

**Methods:**

#### `(j ResponseMapper) Map(ctx context.Context, result any, w http.ResponseWriter) error`

Упаковка результата в SOAP-конверт.

## Functions

####

`DefaultWrapper(logger log.Logger, logMiddleware endpoint.LogMiddleware, restMiddlewares ...http.Middleware) endpoint.Wrapper`

Создает обертку для обработчиков с предустановленными middleware:

- Ограничение размера тела запроса (по умолчанию 64 МБ).
- Добавление RequestId.
- Логирование и метрики.
- Трейсинг.
- Обработка ошибок через ErrorHandler.

#### `ErrorHandler(logger log.Logger) http2.Middleware`

Middleware для перехвата ошибок, их логирования и преобразования в SOAP Fault.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"encoding/xml"
	"log"
	"net/http"

	"github.com/txix-open/isp-kit/http/endpoint/httplog"
	"github.com/txix-open/isp-kit/http/soap"
	log2 "github.com/txix-open/isp-kit/log"
)

type UserRequest struct {
	XMLName xml.Name `xml:"UserRequest"`
	Name    string   `xml:"name"`
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}
	mux := soap.NewActionMux()

	wrapper := soap.DefaultWrapper(logger, httplog.Log(logger, true))
	handler := wrapper.Endpoint(func(ctx context.Context, req UserRequest) {
		/* put here business logic */
	})
	mux.Handle("CreateUser", handler)

	http.ListenAndServe(":8080", mux)
}

```
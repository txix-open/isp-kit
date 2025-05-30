# Package `rabbitmq`

Пакет `rabbitmq` содержит вспомогательные утилиты и адаптеры для взаимодействия с AMQP 0.9.1 (RabbitMQ), в частности для работы с заголовками сообщений (`TableCarrier`) и построения строки назначения (`Destination`).

## Types

### TableCarrier

Тип `TableCarrier` является адаптером типа `amqp091.Table` и реализует интерфейс `TextMapCarrier` из OpenTelemetry.
Он позволяет работать с AMQP-заголовками в виде пары ключ-значение.

**Methods:**

#### `(c TableCarrier) Get(key string) string`

Получить строковое значение по ключу. Если ключ отсутствует, возвращается пустая строка.

#### `(c TableCarrier) Set(key string, value string)`

Установить строковое значение по заданному ключу. Преобразование обратно в тип `interface{}` осуществляется автоматически.

#### `(c TableCarrier) Keys() []string`

Возвращает список всех ключей в `TableCarrier`.

## Functions

### `func Destination(exchange string, routingKey string) string`

Формирует строку назначения на основе переданных `exchange` и `routingKey`. Если `exchange` пустой, возвращается только `routingKey`. В противном случае результат имеет вид `exchange/routingKey`.

## Usage

### Default usage flow

```go
package main

import (
	"fmt"

	"github.com/txix-open/isp-kit/transport/rabbitmq"
)

func main() {
	table := rabbitmq.TableCarrier{}
	table.Set("trace_id", "12345")
	fmt.Println("Trace ID:", table.Get("trace_id"))
	fmt.Println("Keys:", table.Keys())

	dest := rabbitmq.Destination("logs", "info")
	fmt.Println("Destination:", dest)
}
```

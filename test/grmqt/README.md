# Package `grmqt`

Пакет `grmqt` предоставляет вспомогательную обёртку над RabbitMQ-клиентом для функционального тестирования с автоматическим созданием и удалением виртуального хоста, а также публикацией и получением сообщений.

## Types

### Client

Структура `Client` инкапсулирует соединение с RabbitMQ и предоставляет методы для удобной работы с очередями, публикацией сообщений и конфигурацией клиента `grmqx` в тестах.

**Methods:**

#### `func New(t *test.Test) *Client`

Создаёт новый экземпляр клиента. Создаёт виртуальный хост RabbitMQ для теста, настраивает соединение и клиент `grmqx`. Все ресурсы автоматически очищаются при завершении теста.

#### `(c *Client) ConnectionConfig() grmqx.Connection`

Возвращает конфигурацию соединения `grmqx.Connection`, используемую клиентом.

#### `(c *Client) QueueLength(queue string) int`

Возвращает количество сообщений в указанной очереди.

#### `(c *Client) Upgrade(config grmqx.Config)`

Вызывает метод `Upgrade` у клиента `grmqx`, подставляя URL из конфигурации соединения.

#### `(c *Client) PublishJson(exchange string, routingKey string, data any)`

Публикует сообщение в формате JSON в указанный обменник и routing key.

#### `(c *Client) Publish(exchange string, routingKey string, messages ...amqp091.Publishing)`

Публикует одно или несколько сообщений в указанный обменник и routing key.

#### `(c *Client) DrainMessage(queue string) amqp091.Delivery`

Получает одно сообщение из очереди. Тест завершится с ошибкой, если сообщение отсутствует.

#### `(c *Client) useChannel(f func(ch *amqp091.Channel))`

Вспомогательный метод для работы с каналом AMQP. Открывает, передаёт в функцию `f`, а затем закрывает канал.

## Usage

### Default usage flow

```go
package mypkg_test

import (
	"testing"

	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/grpct"
)

func TestPublishAndDrain(t *testing.T) {
	testCtx := test.New(t)
	client := grmqt.New(testCtx)

	client.PublishJson("my-exchange", "my-queue", map[string]string{"hello": "world"})
	msg := client.DrainMessage("my-queue")
	testCtx.Assert().Equal("application/json", msg.ContentType)
}
```

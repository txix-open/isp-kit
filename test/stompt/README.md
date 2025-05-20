# Package `stompt`

Пакет `stompt` предназначен для упрощения написания тестов, использующих STOMP-сообщения через ActiveMQ.

## Types

### Client

Тип `Client` инкапсулирует параметры подключения к STOMP-брокеру и предоставляет методы для генерации конфигураций публикаторов и консьюмеров.

**Methods:**

#### `New(t *test.Test) *Client`

Создаёт новый STOMP-клиент с параметрами подключения из переменных окружения `ACTIVEMQ_STOMP_ADDRESS`, `ACTIVEMQ_USERNAME`, `ACTIVEMQ_PASSWORD`, либо со значениями по умолчанию (`127.0.0.1:61613`, `test`, `test`).

#### `(c *Client) ConsumerConfig(queue string) stompx.ConsumerConfig`

Возвращает конфигурацию STOMP-консьюмера для указанной очереди. Очередь автоматически префиксируется идентификатором теста.

#### `(c *Client) PublisherConfig(queue string) stompx.PublisherConfig`

Возвращает конфигурацию STOMP-публикатора для указанной очереди. Очередь автоматически префиксируется идентификатором теста.

#### `(c *Client) Upgrade(consumers ...consumer.Config)`

Запускает консьюмеров через `stompx.NewConsumerGroup`, автоматически регистрируя `Close()` в `t.Cleanup`. Используется для запуска STOMP-консьюмеров в рамках тестов.

## Usage

### Example usage in test

```go
package mypkg_test

import (
	"testing"
	"github.com/txix-open/isp-kit/stompt"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/stompx/consumer"
)

func TestStompExample(t *testing.T) {
	testCtx := test.New(t)
	client := stompt.New(testCtx)

	cfg := client.ConsumerConfig("example")
	cfg.Handler = consumer.HandlerFunc(func(ctx context.Context, msg consumer.Message) error {
		// Обработка сообщения
		return nil
	})

	client.Upgrade(cfg)
}
```

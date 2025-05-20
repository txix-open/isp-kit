# Package `handler`

Пакет `handler` предоставляет инструменты для обработки сообщений RabbitMQ, включая базовые обработчики, middleware и
управление результатами.

## Types

### Sync

Структура `Sync` реализует обработчик сообщений с поддержкой middleware.

**Methods:**

#### `NewSync(logger log.Logger, adapter SyncHandlerAdapter, middlewares ...Middleware) Sync`

Конструктор синхронного обработчика, принимающий на вход адаптер бизнес-логики, который должен реализовывать интерфейс
`SyncHandlerAdapter`
или быть преобразованным к `SyncHandlerAdapterFunc`, если это функция-обработчик.

Опциональные middleware:

- `Metrics(metricStorage ConsumerMetricStorage) Middleware` – middleware для сбора метрик, регистрирующая время
  обработки, размер сообщения, статусы (Ack/Requeue/Retry/DLQ). Принимает на вход хранилище метрик, реализующее
  интерфейс `ConsumerMetricStorage`.
- `Log(logger log.Logger) Middleware` – логирования событий обработки.

#### `(r Sync) Handle(ctx context.Context, delivery *consumer.Delivery)`

Выполняет обработку сообщения и применяет результат (Ack/Requeue/Retry/MoveToDlq). Логирует ошибки при выполнении
операций с брокером.

## Usage

### Custom adapter

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/grmqx/handler"
	log2 "github.com/txix-open/isp-kit/log"
)

type customHandler struct{}

func (h customHandler) Handle(ctx context.Context, delivery *consumer.Delivery) handler.Result {
	/* put here business logic */
	return handler.Ack()
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	var (
		metricStorage = NewMetricStorage() /* ConsumerMetricStorage interface implementation */
		adapter       customHandler
	)
	syncHandler := handler.NewSync(logger, adapter, []handler.Middleware{
		handler.Metrics(metricStorage),
		handler.Log(logger),
	}...)

	/* handler's call for example */
	delivery := new(consumer.Delivery) /* placeholder for example */
	syncHandler.Handle(context.Background(), delivery)
}

```
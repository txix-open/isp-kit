# Package `batch_handler`

Пакет `batch_handler` предоставляет инструменты для пакетной обработки сообщений RabbitMQ, включая базовые обработчики, middleware и
управление результатами.

## Types

### Handler

Обработчик для пакетной обработки сообщений из очереди RabbitMQ.

**Methods:**

#### `New(adapter SyncHandlerAdapter, purgeInterval time.Duration, maxSize int) *Handler`

Конструктор обработчика. Принимает адаптер бизнес-логики, который должен реализовывать интерфейс `SyncHandlerAdapter`
или быть преобразованным к `SyncHandlerAdapterFunc`, если это функция-обработчик. Также принимает интервал очистки
очереди и максимальный ее размер.

#### `(r *Handler) Handle(ctx context.Context, delivery *consumer.Delivery)`

Добавить сообщение в текущий батч.

#### `(r *Handler) Close()`

Завершить работу обработчика сообщений.

### Sync

Структура `Sync` реализует пакетный обработчик сообщений с поддержкой middleware.

**Methods:**

#### `NewSync(logger log.Logger, adapter SyncHandlerAdapter, middlewares ...Middleware) Sync`

Конструктор синхронного обработчика, принимающий на вход адаптер бизнес-логики, который должен реализовывать интерфейс
`SyncHandlerAdapter`
или быть преобразованным к `SyncHandlerAdapterFunc`, если это функция-обработчик.

Опциональные middleware:

- `Metrics(metricStorage ConsumerMetricStorage) Middleware` – middleware для сбора метрик, регистрирующая время
  обработки, размер сообщения, статусы (Ack/Retry/DLQ). Принимает на вход хранилище метрик, реализующее
  интерфейс `ConsumerMetricStorage`.
- `Log(logger log.Logger) Middleware` – логирования событий обработки.
- `Recovery(logger log.Logger) Middleware` – предотвращает падение сервиса при панике в обработчике, логируя ее в ошибку.

#### `(r Sync) Handle(items []Item)`

Выполняет обработку пакета сообщений и применяет установленный результат (Ack/Retry/MoveToDlq) для каждого из них. 
Логирует ошибки при выполнении операций с брокером.

### Result

Структура `Result` содержит результат обработки сообщения.

### BatchItem

Структура `BatchItem` содержит обрабатываемое сообщение из принятого на обработку пакета, позволяет устанавливать 
и управлять результатом обработки сообщения.

#### `(b *BatchItem) Ack()`

Установка результата сообщения как успешно обработанного.

#### `(b *BatchItem) MoveToDlq(err error)`

Установка результата сообщения как требующего повторной обработки, с логированием ошибки.

#### `(b *BatchItem) Retry(err error)`

Установка результата сообщения как неуспешно обработанного, с отправкой в DLQ и логированием ошибки.

### BatchItem

Структура `BatchItems` содержит принятый на обработку пакет сообщений, позволяет устанавливать
и управлять результатом обработки всех сообщений пакета.

#### `(bs BatchItems) AckAll()`

Установка результата сообщений как успешно обработанных.

#### `(bs BatchItems) MoveToDlqAll(err error)`

Установка результата сообщений как требующих повторной обработки, с логированием ошибки.

#### `(bs BatchItems) RetryAll(err error)`

Установка результата сообщений как неуспешно обработанных, с отправкой в DLQ и логированием ошибки.

## Usage

### Custom adapter

```go
package main

import (
	"log"

	"github.com/txix-open/grmq/consumer"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/grmqx/batch_handler"
	log2 "github.com/txix-open/isp-kit/log"
)

type customHandler struct{}

func (h customHandler) Handle(batch batch_handler.BatchItems) {
	for _, item := range batch {
		/* put here business logic */
		item.Ack()
		/*
			if err != nil {
			    item.MoveToDlq(err)
			}
		*/
	}
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
	handler := batch_handler.NewSync(logger, adapter, []batch_handler.Middleware{
		batch_handler.Metrics(metricStorage),
		batch_handler.Log(logger),
		batch_handler.Recovery(logger),
	}...)

	/* handler's call for example */
	batch := make([]*batch_handler.BatchItem, 0) /* placeholder for example */
	handler.Handle(batch)
}

```
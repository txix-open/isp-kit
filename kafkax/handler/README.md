# Package `handler`

Пакет `handler` предоставляет инструменты для обработки сообщений Kafka с поддержкой middleware, управления результатами
обработки (коммит/ретрай)
и интеграции с метриками/логированием. Предназначен для использования в консумера пакета [`kafkax`](../.).

## Types

### Sync

Основная структура для синхронной обработки сообщений. Обеспечивает:

- Применение цепочки middleware
- Обработку результатов (коммит или повтор)
- Централизованное логирование ошибок

**Methods:**

#### `NewSync(logger log.Logger, adapter SyncHandlerAdapter, middlewares ...Middleware) Sync`

Конструктор синхронного обработчика, принимающий на вход адаптер бизнес-логики, который должен реализовывать интерфейс
`SyncHandlerAdapter`.

Доступные middleware:

- `Metrics(metricStorage ConsumerMetricStorage) Middleware` – сбор метрик времени обработки и размера сообщений,
  количества коммитов и ретраев.
- `Log(logger log.Logger) Middleware` – логирование ключевых событий (успешные коммиты, отправка в ретрай с ошибкой).
- `Recovery() Middleware` – предотвращает падение сервиса при панике в обработчике, преобразуя ее в ошибку.

#### `(r Sync) Handle(ctx context.Context, delivery *consumer.Delivery)`

Выполняет обработку сообщения. Автоматически:

- Вызывает цепочку middleware
- Обрабатывает результат (`Commit` или `Retry`)
- Логирует ошибки коммита

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/handler"
	log2 "github.com/txix-open/isp-kit/log"
)

func processMessage(msg []byte) error {
	/* put here business logic */
	return nil
}

func noopHandler(ctx context.Context, d *consumer.Delivery) handler.Result {
	err := processMessage(d.Source().Value)
	if err != nil {
		return handler.Retry(5*time.Second, err)
	}
	return handler.Commit()
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	adapter := handler.SyncHandlerAdapterFunc(noopHandler)
	syncHandler := handler.NewSync(
		logger,
		adapter,
		handler.Log(logger),
	)

	/* handler's call for example */
	delivery := new(consumer.Delivery) /* placeholder for example */
	syncHandler.Handle(context.Background(), delivery)
}

```
# Package `bgjobx`

Пакет `bgjobx` предоставляет инструменты для работы с фоновыми задачами. Основные компоненты: клиент для работы с
задачами, конфигурация воркеров, сбор метрик и логирование.

Данный пакет использует PostgreSQL для организации очередей задач. Для работы клиента необходимо применить к БД [sql-файл](./migration/20220705201051_bgjob.sql).

## Types

### Client

Клиент для взаимодействия с фоновыми задачами. Воркер обнаруживает новую задачу и вызывает функцию-обработчик. При
успехе – задание помечается выполненным. При ошибке – ретраи с экспоненциальной задержкой. После исчерпания попыток –
перемещение в DLQ.

**Методы:**

#### `NewClient(db DBProvider, logger log.Logger) *Client`

Конструктор клиента

#### `Upgrade(ctx context.Context, workerConfigs []WorkerConfig) error`

Обновить конфигурацию воркеров. Остановить старые воркеры и запустить новые согласно переданным конфигурациям

#### `Enqueue(ctx context.Context, req bgjob.EnqueueRequest) error`

Добавить задачу в очередь

#### `BulkEnqueue(ctx context.Context, list []bgjob.EnqueueRequest) error`

Добавить список задач в очередь

#### `Close()`

Остановить все воркеры

### WorkerConfig

Конфигурация обработчика задач

**Поля:**

- `Queue` - имя очереди
- `Concurrency` - количество параллельных обработчиков
- `PollInterval` - интервал опроса очереди
- `Handle` - функция-обработчик задачи

## Метрики

- Время выполнения задачи
- Количество успешных выполнений
- Количество ретраев
- Количество перемещений в DLQ
- Количество ошибок воркеров

## Стандартный обработчик

#### `NewDefaultHandler(adapter bgjob.Handler, metricStorage handler.MetricStorage) handler.Sync`

Используется для добавления стандартных middleware в функцию-обработчик каждого воркера при создании воркеров.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/txix-open/bgjob"
	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/bgjobx"
	"github.com/txix-open/isp-kit/dbrx"
	"github.com/txix-open/isp-kit/dbx"
)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	db := dbrx.New(application.Logger())
	err = db.Upgrade(application.Context(), dbx.Config{
		Host:     "127.0.0.1",
		Port:     "5432",
		Database: "test",
		Username: "test",
		Password: "test",
	})
	if err != nil {
		log.Fatal(err)
	}

	cli := bgjobx.NewClient(db, application.Logger())
	worker := bgjobx.WorkerConfig{
		Queue:        "test",
		Concurrency:  5,
		PollInterval: 1 * time.Second,
		Handle: bgjob.HandlerFunc(func(ctx context.Context, job bgjob.Job) bgjob.Result {
			/* do some work */
			return bgjob.Complete()
		}),
	}
	err = cli.Upgrade(application.Context(), []bgjobx.WorkerConfig{worker})
	if err != nil {
		log.Fatal(err)
	}

	/* enqueue task */
	err = cli.Enqueue(application.Context(), bgjob.EnqueueRequest{
		Queue: "test",
		Type:  "some-type",
		Arg:   []byte(`{"hello": "world"}`),
	})
}

```
# Package `handler`

Пакет `handler` предоставляет инструменты для обработки результатов выполнения bgjob с использованием middleware.

## Types

### Sync

Структура `Sync` реализует обработчик фоновых задач с поддержкой middleware.

**Methods:**

#### `NewSync(adapter SyncHandlerAdapter, middlewares ...Middleware) Sync`

Конструктор синхронного обработчика, принимающий на вход адаптер бизнес-логики, который должен реализовывать интерфейс
`SyncHandlerAdapter`
или быть преобразованным к `SyncHandlerAdapterFunc`, если это функция-обработчик.

Стандартные Middleware:

- `Metrics(storage MetricStorage) Middleware` – middleware для сбора метрик, регистрирующая время
  обработки. Принимает на вход хранилище метрик, реализующее интерфейс `MetricStorage`.
- `Recovery() Middleware` – предотвращает падение сервиса при панике в обработчике, преобразуя ее в ошибку.
- `RequestId() Middleware` – обеспечивает трассировку, берёт `requestId` из `job.RequestId`.
#### `(r Sync) Handle(ctx context.Context, job bgjob.Job) bgjob.Result`

Выполняет обработку сообщения.

### Mux

Структура `Mux` реализует мультиплексор фоновых задач.

**Methods:**

#### `NewMux() *Mux`

Конструктор мультиплексора.

#### `(m *Mux) Register(jobType string, handler SyncHandlerAdapter) *Mux`

Выполняет регистрацию обработчика задачи по ее типу в мультиплексоре.

#### `(m *Mux) Handle(ctx context.Context, job bgjob.Job) Result`

Выполняет вызов обработчика задачи в зависимости от типа задачи. Если обработчик не зарегистрирован, отправляет задачу в DLQ с ошибкой `bgjob.ErrUnknownType`.

## Functions

### Reschedule(by RescheduleBy, opts ...RescheduleOption) Result

Отдает результат перепланировки задачи:

**Методы перепланировки задач:**

#### ByAfterTime(after time.Duration, currentTime time.Time) AfterTime

Перепланировка задачи на дату спустя указанное время.

#### ByCron(cronExpression string, currentTime time.Time) Cron

Перепланировка задачи по cron-выражению.

**Опции перепланирования:**

#### WithArg(arg []byte) RescheduleOption

Опция для указания аргумента при перепланировании.

## Usage

### Custom handler

```go
package main

import (
    "context"
    "time"

    "github.com/txix-open/bgjob"
    "github.com/txix-open/isp-kit/bgjobx"
    "github.com/txix-open/isp-kit/bgjobx/handler"
)

type customHandler struct{}

func (h customHandler) Handle(ctx context.Context, job bgjob.Job) handler.Result {
  /* put here business logic */
  return handler.Reschedule(handler.ByAfterTime(10*time.Minutes, time.Now()))
}

func main() {
  var (
    metricStorage = NewMetricStorage() /* MetricStorage interface implementation */
    adapter       customHandler
  )

  syncHandler := handler.NewSync(adapter, []handler.Middleware{
    handler.Metrics(metricStorage),
    handler.Recovery(),
    handler.RequestId(),
  }...)

  /* handler's call for example */
  job := new(bgjob.Job) /* placeholder for example */
  syncHandler.Handle(context.Background(), job)
}

```
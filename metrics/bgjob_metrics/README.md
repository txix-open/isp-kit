# Package `bgjob_metrics`

Пакет `bgjob_metrics` предоставляет набор метрик для мониторинга фоновых задач (background jobs), выполняемых через очередь.

Метрики охватывают успешные выполнения, ошибки, повторы, перемещения в DLQ и Продолжительность выполнения задач.

## Types

### Storage

Структура, содержащая все необходимые метрики:

#### `execute_duration_ms`

Продолжительность выполнения задачи.

#### `execute_dlq_count`

Количество задач, попавших в DLQ.

#### `execute_retry_count`

Количество повторных запусков.

#### `execute_success_count`

Количество успешных задач.

#### `worker_error_count`

Количество внутренних ошибок воркера.

**Methods:**

#### `func NewStorage(reg *metrics.Registry) *Storage`

Создаёт экземпляр `Storage`, регистрируя соответствующие метрики в Prometheus.

#### `ObserveExecuteDuration(queue, jobType string, duration time.Duration)`

Фиксирует продолжительность выполнения задачи в миллисекундах. Используется `SummaryVec` с лейблами `queue`, `job_type`.

#### `IncRetryCount(queue, jobType string)`

Увеличивает счётчик повторных запусков задач.

#### `IncDlqCount(queue, jobType string)`

Увеличивает счётчик задач, отправленных в DLQ (dead letter queue).

#### `IncSuccessCount(queue, jobType string)`

Увеличивает счётчик успешно завершённых задач.

#### `IncInternalErrorCount()`

Увеличивает счётчик внутренних ошибок в воркере (без указания очереди или типа задачи).

## Prometheus metrics example

```
# HELP bgjob_execute_duration_ms The latency of execution single job from queue
# TYPE bgjob_execute_duration_ms summary
bgjob_execute_duration_ms{queue="billing",job_type="recalc"} 123.4

# HELP bgjob_execute_dlq_count Count of jobs moved to DLQ
# TYPE bgjob_execute_dlq_count counter
bgjob_execute_dlq_count{queue="billing",job_type="recalc"} 3

# HELP bgjob_execute_retry_count Count of retried jobs
# TYPE bgjob_execute_retry_count counter
bgjob_execute_retry_count{queue="billing",job_type="recalc"} 2

# HELP bgjob_execute_success_count Count of successful jobs
# TYPE bgjob_execute_success_count counter
bgjob_execute_success_count{queue="billing",job_type="recalc"} 5

# HELP bgjob_worker_error_count Count of internal worker errors
# TYPE bgjob_worker_error_count counter
bgjob_worker_error_count 1
```

## Usage

### Default usage flow

```go
metrics := bgjob_metrics.NewStorage(metrics.DefaultRegistry)
metrics.ObserveExecuteDuration("billing", "recalc", time.Millisecond*120)
metrics.IncRetryCount("billing", "recalc")
metrics.IncSuccessCount("billing", "recalc")
metrics.IncDlqCount("billing", "recalc")
metrics.IncInternalErrorCount()
```

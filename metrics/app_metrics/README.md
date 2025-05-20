# Package `app_metrics`

Пакет `app_metrics` предоставляет метрики, связанные с логированием приложения. В частности, позволяет отслеживать количество логов, которые были **сэмплированы** или **отброшены** логгером `zap`.

## Types

### LogCounter

Структура, содержащая счётчики для каждого уровня логирования (`log.Level`).

**Metrics:**

#### `app_logs_sampling_status_count`

Количество семплированных и дропнутых логов по `log.Level`.

**Methods:**

#### `func NewLogCounter(registry *metrics.Registry) LogCounter`

Создаёт и регистрирует счётчики Prometheus для логов с лейблами:

- `level` — уровень логирования (`debug`, `info`, `warn`, `error`, `fatal`)
- `status` — `sampled` или `dropped`

Возвращает `LogCounter` с мапой счётчиков по уровням.

#### `func (c LogCounter) SampledLogCounter() func(entry zapcore.Entry) error`

Функция для регистрации в `zapcore.Core.WithOptions(...)`, увеличивающая метрику `sampled` при каждом записанном логе.

#### `func (c LogCounter) DroppedLogCounter() func(zapcore.Entry, zapcore.SamplingDecision)`

Функция для передачи в `zapcore.NewSampler(...)` в качестве обработчика событий, увеличивающая счётчик `dropped` при отбрасывании логов.

## Internal types

### `keeper`

Внутренний тип, содержащий два счётчика:

- `sampled` — записанные (сохранённые) логи
- `dropped` — отброшенные логи

Используется в `map[log.Level]keeper` внутри `LogCounter`.

## Prometheus metrics example

```
# HELP app_logs_sampling_status_count Count of logs statuses(dropped or sampled)
# TYPE app_logs_sampling_status_count counter
app_logs_sampling_status_count{level="info",status="sampled"} 123
app_logs_sampling_status_count{level="debug",status="dropped"} 45
```

## Usage

### Register

```go
logger := zap.New(
    zapcore.NewSampler(
        core, // zapcore.Core
        time.Second,
        100,
        10,
    ).WithOptions(
        zap.WrapCore(func(core zapcore.Core) zapcore.Core {
            counter := app_metrics.NewLogCounter(metrics.DefaultRegistry)
            return zapcore.RegisterHooks(core,
                counter.SampledLogCounter(),
                counter.DroppedLogCounter(),
            )
        }),
    ),
)
```

# Package `sql_tracing`

Пакет `sql_tracing` предоставляет инструменты трассировки SQL-запросов для PostgreSQL-драйвера `pgx` с использованием OpenTelemetry. Поддерживает автоматическое создание span'ов, запись аргументов и SQL-выражений, а также метрики SQL-операций через `sql_metrics`.

## Types

### Tracer

Структура `Tracer` реализует интерфейс `pgx.QueryTracer` и обеспечивает создание и завершение span'ов при выполнении SQL-запросов. Позволяет гибко настраивать включение SQL-выражений и параметров в span'ы.

**Methods:**

#### `NewTracer(tracer trace.Tracer, config Config) Tracer`

Создаёт трейсер.

#### `(t Tracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context`

Создаёт span перед выполнением SQL-запроса. Возвращает новый `context.Context` с активным span.

#### `(t Tracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData)`

Завершает span после выполнения SQL-запроса. В случае ошибки (кроме `sql.ErrNoRows`) записывает её в span и устанавливает статус `Error`.

### Config

Структура `Config` содержит параметры конфигурации для трассировки SQL-запросов.

**Fields:**

#### `Provider tracing.TracerProvider`

Провайдер трассировки (по умолчанию `tracing.DefaultProvider`)

#### `EnableStatement bool`

Включает добавление SQL-выражения в атрибуты span'а

#### `EnableArgs bool`

Включает добавление аргументов запроса в атрибуты span'а

**Methods**

#### `NewConfig() Config`

Создаёт конфигурацию трассировки с провайдером по умолчанию.

#### `(c Config) QueryTracer() pgx.QueryTracer`

Возвращает либо полноценный трассировщик `Tracer`, либо `noop`, если провайдер трассировки является "заглушкой".

## Usage

### Подключение трассировки SQL-запросов

```go
import (
    "github.com/txix-open/isp-kit/observability/tracing"
    "github.com/txix-open/isp-kit/observability/sql_tracing"

    "github.com/jackc/pgx/v5"
)

func setupTracer() pgx.QueryTracer {
    cfg := sql_tracing.NewConfig()
    cfg.EnableStatement = true
    cfg.EnableArgs = true

    return cfg.QueryTracer()
}
```

### Регистрация в `pgx` соединении

```go
connConfig := pgx.ConnConfig{ /* ваша конфигурация */ }
connConfig.Tracer = setupTracer()
```

Теперь каждый SQL-запрос будет автоматически трассироваться с учётом включённых опций (запрос, параметры, request ID и т.п.).

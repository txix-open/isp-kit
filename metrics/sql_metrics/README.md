# Package `sql_metrics`

Пакет `sql_metrics` предоставляет инструменты для сбора метрик времени выполнения SQL-запросов с использованием Prometheus и клиента PostgreSQL `pgx`.

## Types

### `QueryDurationMetrics`

Хранилище метрик времени выполнения SQL-запросов.

#### `func NewTracer(reg *metrics.Registry) QueryDurationMetrics`

Создаёт новое хранилище метрик SQL-запросов.

**Metrics:**

#### `sql_query_duration_ms`

Продолжительность выполнения SQL-запроса, метрика `summary` с лейблом `operation`.

**Methods:**

#### `func (m QueryDurationMetrics) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context`

Сохраняет время начала выполнения запроса, если в контексте установлен лейбл операции.

#### `func (m QueryDurationMetrics) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData)`

Фиксирует задержку выполнения SQL-запроса и регистрирует её в метриках. Использует лейбл из контекста.

#### `func OperationLabelToContext(ctx context.Context, label string) context.Context`

Сохраняет метку операции SQL-запроса в контекст.

#### `func OperationLabelFromContext(ctx context.Context) string`

Извлекает метку операции SQL-запроса из контекста.

## Prometheus metrics example

```
# HELP sql_query_duration_ms The latencies of sql query
# TYPE sql_query_duration_ms summary
sql_query_duration_ms{operation="get_user"} 12.3
sql_query_duration_ms{operation="create_order"} 25.7
```

## Usage

### Default usage flow

```go
tracer := sql_metrics.NewTracer(metrics.DefaultRegistry)

ctx := sql_metrics.OperationLabelToContext(context.Background(), "get_user")

config.Tracer = &pgx.Tracer{ // при использовании pgx
    TraceQueryStart: tracer.TraceQueryStart,
    TraceQueryEnd: tracer.TraceQueryEnd,
}
```

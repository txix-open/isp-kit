# Package `pgx_metrics`

Пакет `pgx_metrics` предоставляет функцию для регистрации и публикации метрик состояния пула подключений PostgreSQL на базе `pgxpool`.

## Types

Данный пакет не экспортирует пользовательских типов.

## Functions

### `func Register(reg *metrics.Registry, statsProvider statsProvider, dbName string)`

Регистрирует метрики состояния `pgxpool.Pool` в Prometheus-реестре.

## Prometheus metrics example

```
# HELP pgxpool_acquire_count Cumulative count of successful acquires from the pool.
# TYPE pgxpool_acquire_count counter

# HELP pgxpool_acquire_duration_seconds_total Total time spent acquiring connections (nanoseconds in source; should be converted to seconds).
# TYPE pgxpool_acquire_duration_seconds_total counter

# HELP pgxpool_acquired_conns Number of currently acquired connections in the pool.
# TYPE pgxpool_acquired_conns gauge

# HELP pgxpool_canceled_acquire_count Cumulative count of acquires from the pool that were canceled by a context.
# TYPE pgxpool_canceled_acquire_count counter

# HELP pgxpool_constructing_conns Number of connections with construction in progress in the pool.
# TYPE pgxpool_constructing_conns gauge

# HELP pgxpool_empty_acquire_count Cumulative count of successful acquires that waited due to empty pool.
# TYPE pgxpool_empty_acquire_count counter

# HELP pgxpool_empty_acquire_wait_seconds_total Total time waiting for successful acquires due to empty pool (nanoseconds in source; should be converted to seconds).
# TYPE pgxpool_empty_acquire_wait_seconds_total counter

# HELP pgxpool_idle_conns Number of currently idle connections in the pool.
# TYPE pgxpool_idle_conns gauge

# HELP pgxpool_max_conns Maximum size of the pool.
# TYPE pgxpool_max_conns gauge

# HELP pgxpool_total_conns Total number of resources currently in the pool (constructing + acquired + idle).
# TYPE pgxpool_total_conns gauge

# HELP pgxpool_new_conns_count Cumulative count of new connections opened.
# TYPE pgxpool_new_conns_count counter

# HELP pgxpool_max_lifetime_destroy_count Cumulative count of connections destroyed due to MaxConnLifetime.
# TYPE pgxpool_max_lifetime_destroy_count counter

# HELP pgxpool_max_idle_destroy_count Cumulative count of connections destroyed due to MaxConnIdleTime.
# TYPE pgxpool_max_idle_destroy_count counter
```

## Usage

### Default usage flow

```go
import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/pgx_metrics"
)

func setupDB(ctx context.Context) {
	pool, _ := pgxpool.New(ctx, "postgres://...")

	pgx_metrics.Register(
		metrics.DefaultRegistry,
		pool,
		"main",
	)
}
```
# Package `pgx_metrics`

Пакет `pgx_metrics` предоставляет функцию для регистрации и публикации метрик состояния пула подключений PostgreSQL на базе `pgxpool`.

## Types

Данный пакет не экспортирует пользовательских типов.

## Functions

### `func Register(reg *metrics.Registry, statsProvider statsProvider, dbName string)`

Регистрирует метрики состояния `pgxpool.Pool` в Prometheus-реестре.

## Prometheus metrics example

```
# HELP pgxpool_acquired_conns Current number of acquired connections in the pool.
# TYPE pgxpool_acquired_conns gauge

# HELP pgxpool_idle_conns Current number of idle connections in the pool.
# TYPE pgxpool_idle_conns gauge

# HELP pgxpool_total_conns Total number of connections in the pool.
# TYPE pgxpool_total_conns gauge

# HELP pgxpool_max_conns Maximum number of connections allowed in the pool.
# TYPE pgxpool_max_conns gauge

# HELP pgxpool_constructing_conns Number of connections currently being created.
# TYPE pgxpool_constructing_conns gauge

# HELP pgxpool_new_conns New connections being opened.
# TYPE pgxpool_new_conns gauge

# HELP pgxpool_acquire_count Total number of successful connection acquires.
# TYPE pgxpool_acquire_count counter

# HELP pgxpool_canceled_acquire_count Number of acquires canceled by context.
# TYPE pgxpool_canceled_acquire_count counter

# HELP pgxpool_empty_acquire_count Number of acquires that waited due to empty pool.
# TYPE pgxpool_empty_acquire_count counter

# HELP pgxpool_acquire_duration_seconds_total Total time spent acquiring connections.
# TYPE pgxpool_acquire_duration_seconds_total counter

# HELP pgxpool_empty_acquire_wait_seconds_total Total time waiting for connections when pool is empty.
# TYPE pgxpool_empty_acquire_wait_seconds_total counter
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
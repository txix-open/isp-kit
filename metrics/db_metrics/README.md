# Package `db_metrics`

Пакет `db_metrics` предоставляет вспомогательную функцию для регистрации метрик состояния подключения к базе данных с использованием встроенного Prometheus-коллектора `DBStatsCollector`.

## Types

Данный пакет не экспортирует пользовательских типов.

## Functions

### `func Register(reg *metrics.Registry, db *sql.DB, dbName string)`

Регистрирует метрики состояния `sql.DB` в Prometheus-реестре.

## Prometheus metrics example

Метрики, регистрируемые `DBStatsCollector`, включают (но не ограничиваются):

```
# HELP go_sql_db_max_open_connections Maximum number of open connections to the database.
# TYPE go_sql_db_max_open_connections gauge
# HELP go_sql_db_open_connections The number of established connections both in use and idle.
# TYPE go_sql_db_open_connections gauge
# HELP go_sql_db_in_use_connections The number of connections currently in use.
# TYPE go_sql_db_in_use_connections gauge
# HELP go_sql_db_idle_connections The number of idle connections.
# TYPE go_sql_db_idle_connections gauge
# HELP go_sql_db_wait_count The total number of connections waited for.
# TYPE go_sql_db_wait_count counter
# HELP go_sql_db_wait_duration_seconds The total time blocked waiting for a new connection.
# TYPE go_sql_db_wait_duration_seconds counter
# HELP go_sql_db_max_idle_closed The total number of connections closed due to SetMaxIdleConns.
# TYPE go_sql_db_max_idle_closed counter
# HELP go_sql_db_max_lifetime_closed The total number of connections closed due to SetConnMaxLifetime.
# TYPE go_sql_db_max_lifetime_closed counter
```

## Usage

### Default usage flow

```go
import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/db_metrics"
)

func setupDB() {
	db, _ := sql.Open("postgres", "...")
	db_metrics.Register(metrics.DefaultRegistry, db, "main")
}
```

После этого все стандартные метрики состояния пулов соединений будут автоматически отображаться через `/metrics` endpoint.

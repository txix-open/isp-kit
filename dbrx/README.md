# Package `dbrx`

Пакет `dbrx` предоставляет функциональность для работы с базой данных PostgreSQL через расширенный клиент,
поддерживающий динамическое обновление конфигурацию, интеграцию с метриками и трейсингом, healthcheck для мониторинга,
а также управление жизненным циклом подключения с безопасным обновлением соединений.

## Types

### Client

Клиент для взаимодействия с базой данных PostgreSQL, расширяющий клиент из [пакета `dbx`](../dbx/db.go).

**Methods:**

Содержит все методы клиента-предшественника.

#### `New(logger log.Logger, opts ...dbx.Option) *Client`

Создание клиента к базе данных с пустой конфигурацией подключения. По-умолчанию данный клиент автоматически создает
схему БД. Доступные опции:

- `WithQueryTracer(tracers ...pgx.QueryTracer) Option` – трассировка запросов с объектами реализующими интерфейс
  `pgx.QueryTracer`.
- `WithMigrationRunner(migrationDir string, logger log.Logger) Option` – применить sql-скрипты для миграции, находящиеся
  в указанной директории.
- `WithCreateSchema(createSchema bool) Option` – включить/выключить функцию создания схемы БД.

#### `(c *Client) Upgrade(ctx context.Context, config dbx.Config) error`

Обновить конфигурацию подключения к базе данных.

#### `(c *Client) DB() (*dbx.Client, error)`

Получить родительский клиент из [пакета `dbx`](../dbx/db.go).

#### `(c *Client) Healthcheck(ctx context.Context) error`

Проверить доступность соединения с бд.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/dbrx"
	"github.com/txix-open/isp-kit/dbx"
	log2 "github.com/txix-open/isp-kit/log"
)

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}
	cli := dbrx.New(logger)

	ctx := context.Background()
	cfg := dbx.Config{
		Host:     "127.0.0.1",
		Port:     "5432",
		Database: "test",
		Username: "test",
		Password: "test",
		Schema:   "test",
	}
	err = dbrx.Upgrade(ctx, cfg)
	if err != nil {
		log.Fatal(err)
    }

	var result int
	err = cli.SelectRow(ctx, &result, "SELECT COUNT(*) FROM users")
	if err != nil {
		log.Fatal(err)
	}
}

```
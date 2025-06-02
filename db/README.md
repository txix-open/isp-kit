# Package `db`

Пакет `db` предоставляет клиент-обёртку над библиотекой [`sqlx`](https://github.com/jmoiron/sqlx) с интеграцией `pgx` для работы с PostgreSQL. Реализует
функциональность подключения к базе данных, выполнения запросов, транзакций и настройки трассировки запросов

## Types

### Client

Основная структура-клиент для взаимодействия с базой данных PostgreSQL

**Methods:**

Содержит все методы клиента-предшественника из [`sqlx`](https://github.com/jmoiron/sqlx).

#### `Open(ctx context.Context, dsn string, opts ...Option) (*Client, error)`

Создать подключение к базе данных по переданному dsn. Доступные опции:
- `WithQueryTracer(tracers ...pgx.QueryTracer) Option` – трассировка запросов с объектами реализующими интерфейс `pgx.QueryTracer`.

#### `(db *Client) Select(ctx context.Context, ptr any, query string, args ...any) error`

Выполнить sql-запрос с возвращением данных из БД. Использует стандартные плейсхолдеры для позиционирования параметров.

#### `(db *Client) SelectRow(ctx context.Context, ptr any, query string, args ...any) error`

Выполнить sql-запрос с возвращением одной строки данных из БД. Использует стандартные плейсхолдеры для позиционирования параметров.

#### `(db *Client) Exec(ctx context.Context, query string, args ...any) (sql.Result, error)`

Выполнить sql-запрос без возврата данных из БД. Использует стандартные плейсхолдеры для позиционирования параметров.

#### `(db *Client) ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error)`

Выполнить sql-запрос без возврата данных из БД. Позволяет использовать имена полей вместо плейсхолдеров для позиционирования параметров в sql-запросе.

#### `(db *Client) RunInTransaction(ctx context.Context, txFunc TxFunc, opts ...TxOption) (err error)`

Выполнить функцию `txFunc` в рамках транзакции. Доступны опции:
- `IsolationLevel(level sql.IsolationLevel) TxOption` – изменить уровень изоляции транзакции
- `ReadOnly() TxOption` – режим транзакции только на чтение

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"database/sql"
	"log"

	"github.com/txix-open/isp-kit/db"
)

type user struct {
	Name     string
	IsActive bool
}

func main() {
	ctx := context.Background()
	client, err := db.Open(ctx, "postgres://user:pass@localhost:5432/db")
	if err != nil {
		log.Fatal(err)
	}

	var users []user
	err = client.Select(ctx, &users, "SELECT * FROM users WHERE is_active = $1", true) /* read data */

	/* transaction */
	err = client.RunInTransaction(ctx, func(ctx context.Context, tx *db.Tx) error {
		res, err := tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1", 100)
		if err != nil {
			return err
		}
		/* put here some business logic */
		return nil
	}, db.IsolationLevel(sql.LevelRepeatableRead))
	if err != nil {
		log.Fatal(err)
	}
}

```
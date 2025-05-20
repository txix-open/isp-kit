# Package `dbt`

Пакет `dbt` предназначен для создания и управления изолированной тестовой базой данных PostgreSQL с уникальной схемой на каждый запуск тестов.

## Types

### TestDb

Структура `TestDb` инкапсулирует клиента базы данных `dbx.Client`, используемого для работы с тестовой схемой, а также предоставляет удобные методы для выполнения SQL-запросов с автоматической проверкой ошибок.

**Methods:**

#### `New(t *test.Test, opts ...dbx.Option) *TestDb`

Создаёт новое подключение к базе данных с уникальной схемой, которая будет автоматически удалена по завершению теста. Внутри создаётся контекст с таймаутом, извлекаемым из конфигурации (`PG_OPEN_TIMEOUT`).

#### `(db *TestDb) DB() (*dbx.Client, error)`

Возвращает клиент `dbx.Client`.

#### `(db *TestDb) Must() must`

Возвращает структуру `must`, предоставляющую обёртки над методами выполнения SQL-запросов с автоматическими ассерциями.

#### `(db *TestDb) Schema() string`

Возвращает имя схемы, с которой работает экземпляр `TestDb`.

#### `(db *TestDb) Close() error`

Удаляет схему из базы данных (`DROP SCHEMA ... CASCADE`) и закрывает подключение. В случае ошибок оборачивает их с дополнительным контекстом.

### must

Структура `must` предоставляет методы для безопасного выполнения SQL-запросов с проверками через `require.Assertions`. Используется исключительно в тестах.

**Methods:**

#### `(m must) Exec(query string, args ...any) sql.Result`

Выполняет SQL-запрос и проверяет, что ошибка отсутствует.

#### `(m must) Select(resultPtr any, query string, args ...any)`

Выполняет SQL-запрос, ожидающий множество строк, и проверяет отсутствие ошибки.

#### `(m must) SelectRow(resultPtr any, query string, args ...any)`

Выполняет SQL-запрос, ожидающий одну строку, и проверяет отсутствие ошибки.

#### `(m must) ExecNamed(query string, arg any) sql.Result`

Выполняет именованный SQL-запрос и проверяет отсутствие ошибки.

#### `(m must) Count(query string, args ...any) int`

Выполняет SQL-запрос, возвращающий одно числовое значение (count), и проверяет отсутствие ошибки.

### `Config(t \*test.Test) dbx.Config`

Формирует конфигурацию для подключения к базе данных на основе значений из тестовой конфигурации. Создаёт уникальное имя схемы (`test_<id>`) для каждого запуска теста.

## Usage

### Default usage flow

```go
package mypkg_test

import (
	"testing"

	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
)

func TestExample(t *testing.T) {
	ctx := context.Background()
	testCtx := test.New(t)
	db := dbt.New(testCtx)

	// Пример использования метода must.Exec
	db.Must().Exec(`CREATE TABLE example (id SERIAL PRIMARY KEY, name TEXT)`)
	db.Must().Exec(`INSERT INTO example (name) VALUES ($1)`, "test")

	var count int
	db.Must().SelectRow(&count, `SELECT COUNT(*) FROM example`)
	testCtx.Assert().Equal(1, count)

	// Очистка схемы произойдёт автоматически по завершению
}
```

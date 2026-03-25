# Package `log`

Пакет `log` предоставляет функциональность для структурированного логирования с интеграцией в контекст приложения.
Основан на [zap](https://github.com/uber-go/zap) с расширенными возможностями.

## Types

### Adapter

Структура `RoundRobin` реализует алгоритм балансировки нагрузки Round Robin.

**Methods:**

#### `New(opts ...Option) (*Adapter, error)`

Конструктор логгера. Поддерживает следующие опции:

- `WithDevelopmentMode() Option` – включить логирование в режиме разработки
- `WithFileOutput(fileOutput file.Output) Option` – добавить запись логов в файл
- `WithLevel(level Level) Option` – изменить уровень логирования

#### `NewFromConfig(config Config) (*Adapter, error)`

Конструктор логгера через объект конфигурации.

#### `(a *Adapter) Fatal(ctx context.Context, message any, fields ...Field)`

Логирование сообщения с уровнем fatal.

#### `(a *Adapter) Error(ctx context.Context, message any, fields ...Field)`

Логирование сообщения с уровнем error.

#### `(a *Adapter) Warn(ctx context.Context, message any, fields ...Field)`

Логирование сообщения с уровнем warn.

#### `(a *Adapter) Info(ctx context.Context, message any, fields ...Field)`

Логирование сообщения с уровнем info.

#### `(a *Adapter) Debug(ctx context.Context, message any, fields ...Field)`

Логирование сообщения с уровнем debug.

#### `(a *Adapter) Log(ctx context.Context, level Level, message any, fields ...Field)`

Логирование сообщения с указанным уровнем `level`.

#### `(a *Adapter) SetLevel(level Level)`

Установить указанный уровень логирования.

#### `(a *Adapter) Enabled(level Level) bool`

Проверить активность указанного уровня логирования.

#### `(a *Adapter) Sync() error`

Синхронизация буферов логера.

#### `(a *Adapter) Config() Config`

Получить конфиг логера.

## Functions

#### `StdLoggerWithLevel(adapter Logger, level Level, withFields ...Field) *log.Logger`

Преобразовать логера из текущего пакета в логер из стандартной библиотеки `log`

#### `ContextLogValues(ctx context.Context) []Field`

Получить поля для логов из контекста.

#### `ToContext(ctx context.Context, kvs ...Field) context.Context`

Добавить поля для логов в контекст.

#### `RewriteContextField(ctx context.Context, field Field) context.Context`

Перезаписать поле для логов в контексте.

#### `UpsertContextField(ctx context.Context, field Field) context.Context`

Перезаписать или записать поле для логов в контексте.

#### `FromContext(ctx context.Context, key string) (Field, bool)`

Получить поле для логов по ключу, если не найдено ok = false

#### `CopyValues(ctxTo, ctxFrom context.Context) context.Context`

Копирует в контекст все поля второго контекста  

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	log2 "github.com/txix-open/isp-kit/log"
)

func main() {
	logger, err := log2.New(log2.WithLevel(log2.InfoLevel))
	if err != nil {
		log.Fatal(err)
	}

	/* store log fields in context */
	ctx := log2.ToContext(context.Background(),
		log2.String("requestId", "ff4488"),
		log2.String("secret", "x-secret-key"),
	)

	logger.Info(ctx, "hello world!",
		log2.String("service", "greetings-service"),
	)
	
	oldRequestId, ok := log2.FromContext(ctx, "requestId")
	if ok {
		ctx = log2.UpsertContextField(ctx, log2.String("requestId", "aa2266"))
		logger.Info(ctx, "new and old requestId",
			oldRequestId,
		)
    }
	
	

	/* change logger's level dynamically */
	logger.SetLevel(log2.DebugLevel)
	logger.Debug(ctx, "log level changed to debug")

	err = logger.Sync()
	if err != nil {
		log.Fatal(err)
	}
}

```
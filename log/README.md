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
- `WithSplit(maxLineKb int) Option` – включить разбиение больших JSON-записей на части

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

## Разбиение больших записей на части

Если одна JSON-запись не проходит в лимит лог-коллектора,
её можно писать несколькими обычными записями. Разбиение включается полем `Split` в
`log.Config` или опцией `log.WithSplit(maxLineKb)`, в `bootstrap` — секцией `logs.split`
локального конфига:

```yaml
logs:
  split:
    enable: true    # по умолчанию false
    maxLineKb: 218   # размер одной части в килобайтах, 218 × 1024 = 223 232 байта
```

Ключи конфига: `LOGS.SPLIT.ENABLE` (по умолчанию `false`) и `LOGS.SPLIT.MAXLINEKB`
(по умолчанию `218`). Одного `maxLineKb` без `enabled: true` недостаточно. В dev-режиме
разбиение не работает: вывод не JSON.

### Контракт части

Запись, которая не влезает в `maxLineKb × 1024` байт, превращается в несколько **обычных**
лог-записей: у каждой свой `level`, `ts` и исходные имена полей. Ничего не
переименовывается и не заворачивается в `payload`. В корень каждой части добавляются:

* `logPackageId` — UUID v4 (RFC 9562), общий для всех частей одной исходной записи;
* `logPackageIndex` — порядковый номер части, начиная с `1`;
* `logPackageTotal` — общее количество частей.

`requestId`, если он был в записи, попадает в корень **каждой** части. `level`, `ts`, `msg` и
`requestId` одинаковы во всех частях: по ним можно фильтровать в VL, и сообщение не теряется
ни в одной части.

Мелкие поля упаковываются в части жадно в исходном порядке. Поле, которое само не влезает в
часть, остаётся под своим именем, а его значение режется на куски: части идут в порядке
`logPackageIndex`, после сборки значение восстанавливается склейкой.

Пример: `responseBody` не влезает в лимит.

```json
{"level":"debug","ts":"...","msg":"http handler: response","requestId":"3365548a...",
 "logPackageId":"9f2c...","logPackageIndex":1,"logPackageTotal":3,
 "statusCode":200,"elapsedTimeMs":302}
{"level":"debug","ts":"...","msg":"http handler: response","requestId":"3365548a...",
 "logPackageId":"9f2c...","logPackageIndex":2,"logPackageTotal":3,
 "responseBody":"{\"added\":[],\"changed\":[],\"removed\":[{\"block\":\"ADDRESS\"..."}
{"level":"debug","ts":"...","msg":"http handler: response","requestId":"3365548a...",
 "logPackageId":"9f2c...","logPackageIndex":3,"logPackageTotal":3,
 "responseBody":"...\"errorType\":\"Уникальность связки\",...}]}"}
```

### Сборка на приёмнике

```
parts := group by logPackageId, sort by logPackageIndex
if len(parts) != logPackageTotal: пачка битая — отбросить

result := {}
for key in {level, ts, msg, requestId}:
    if key in parts[0]:
        result[key] = parts[0][key]   // одинаковые во всех частях, берём из любой

values := map[string][]any    // значения полей по частям, в порядке logPackageIndex

for part in parts:
    for key, value in part:
        if key in {level, ts, msg, logPackageId, logPackageIndex, logPackageTotal, requestId}:
            continue
        values[key] = append(values[key], value)

for key, list in values:
    if len(list) == 1:
        result[key] = list[0]     // обычное поле (число, строка, bool)
    else:
        result[key] = concat(list) // куски строкового поля, склеиваем по порядку
```

Правило простое: поле, которое встречается больше одного раза, — это куски одного значения;
встречается один раз — обычное поле как есть. `requestId` и `msg` повторяются в каждой части,
но значения одни и те же, поэтому берутся из любой части.

### Ограничения

* Разбиваются только строковые поля (`String`, `ByteString`, `Error`, `fmt.Stringer`). Если в лимит не влезает
  нестроковое поле (объект, массив, число), разбиение не применяется и запись пишется как есть.
* `Any` разбивается по типу значения: `string`, `[]byte`, `error` и `fmt.Stringer` — как соответствующее строковое поле 
  (zap резолвит их в конкретный тип при создании поля). Значения `struct`, `map`, `slice` (в том числе `[]string`) кодируются как JSON-объект/массив и не разбиваются:
  куски склеиваются конкатенацией, для объектов и массивов восстановление структуры невозможно. Если такой объект не влезает в лимит, запись пишется как есть.
  Большое тело (например, JSON-ответ) логируйте строкой — `log.String` или `log.ByteString`.
* Если в лимит не влезает сам `msg`, запись пишется как есть.
* Невалидный UTF-8 в большом поле не режется: ширину экранирования нельзя посчитать точно,
  запись пишется как есть.
* Поля с пространствами имён (`zap.Namespace`) не разбиваются: после открытия namespace
  последующие поля вложены в него, поэтому запись пишется как есть.
* Поля, добавленные к логгеру через `With` (в том числе в `StdLoggerWithLevel`), попадают в
  каждую часть; они учитываются при расчёте бюджета, но не режутся.
* Нарезка идёт по байтам с учётом экранирования и откатом до начала UTF-8-символа, поэтому
  кириллица и эмодзи не рвутся, а каждая часть укладывается в `maxLineKb × 1024` байт.
* Сэмплирование принимает решение по исходной записи один раз, поэтому пачка пишется целиком
  или не пишется вовсе; хуки вызываются один раз на запись.
* `level`, `ts`, `msg` и `requestId` дублируются в каждой части (~200 байт накладных на часть):
  это плата за то, что часть остаётся полноценной лог-записью и по ней можно фильтровать.
* Если запись не удалось разложить, она пишется как есть — сбой разбиения не теряет лог.

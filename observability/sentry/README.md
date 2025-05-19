# Package `sentry`

Пакет `sentry` предоставляет интеграцию с Sentry SDK для централизованного сбора и отправки ошибок и событий, с возможностью обогащения контекста через `context.Context` и функции обогащения событий.

## Types

### Config

Конфигурация для sentry

**Fields:**

#### **Enable**

Флаг включения Sentry.

#### **Dsn**

DSN Sentry.

#### **ModuleName**

Имя сервиса/модуля.

#### **ModuleVersion**

Версия модуля.

#### **Environment**

Окружение (например, `prod`, `dev`).

#### **InstanceId**

Уникальный ID инстанса.

#### **Tags**

Дополнительные произвольные теги.

### Hub

Интерфейс `Hub` для отправки ошибок и событий:

**Methods:**

#### `CatchError(ctx context.Context, err error, level log.Level)`

Отправить ошибку с уровнем логирования.

#### `CatchEvent(ctx context.Context, event *sentry.Event)`

Отправить события напрямую.

#### `Flush()`

Подождать отправки всех событий.

### EventEnrichment

Функция обогащения события — добавляет произвольные данные в `sentry.Event`.

## Logger

- Проксирует все методы логгера.
- Автоматически отправляет события в Sentry, если уровень лога присутствует в `supportedLevels`.
- Добавляет в событие:

  - уровень логирования,
  - `requestId` из контекста (если есть),
  - ошибку (если передана в лог),
  - данные обогащения из контекста.

## Реализации Hub

### SdkHub

Реальная обёртка над `sentry.Hub`. Отправляет события в Sentry SDK.

### NoopHub

Заглушка, которая ничего не делает. Используется, если Sentry отключён.

## Functions

### `func NewHubFromConfiguration(config Config) (Hub, error)`

Создаёт `Hub` на основе конфигурации. Возвращает `NoopHub`, если `Enable == false`.

### `func EnrichEvent(ctx context.Context, enrichment EventEnrichment) context.Context`

Добавляет функцию обогащения события в контекст. Используется для динамического расширения `sentry.Event` в `Logger`.

### `func WrapLogger(logger log.Logger, hub Hub, supportedLevels []log.Level) Logger`

Оборачивает логгер, чтобы автоматически отправлять события в Sentry на указанных уровнях (`supportedLevels`).

### `func WrapErrorLogger(logger log.Logger, hub Hub) Logger`

Упрощённый вариант `WrapLogger`, логирует только ошибки (`log.LevelError`).

## Usage

### Оборачивание логгера

```go
wrappedLogger := sentry.WrapLogger(baseLogger, hub, []log.Level{log.LevelError, log.LevelFatal})
```

### Обогащение событий

```go
ctx = sentry.EnrichEvent(ctx, func(event *sentry.Event) {
	event.Extra["user_id"] = 12345
})
```

### Использование с контекстом

```go
hub.CatchError(ctx, err, log.LevelError)
```

Ошибка будет автоматически дополнена `requestId` и пользовательскими полями из `EnrichEvent`.

---

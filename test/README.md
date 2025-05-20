# Package `test`

Пакет `test` предоставляет обёртку над `testing.T`, а также автоматическую инициализацию конфигурации и логгера для использования в юнит- и интеграционных тестах.

## Types

### Test

Структура `Test` служит основным контейнером для всех зависимостей, часто используемых в тестах: конфигурация, логгер, контекст `testing.T`, а также вспомогательные утверждения `require.Assertions`.

**Fields:**

#### `id string`

Случайно сгенерированный идентификатор теста (hex-строка длиной 8 символов).

#### `cfg *config.Config`

Конфигурация, созданная с помощью `config.New()`.

#### `logger log.Logger`

Логгер, инициализированный в режиме разработки и с уровнем `Debug`.

#### `t *testing.T`

Ссылка на контекст текущего теста.

#### `assertions *require.Assertions`

Обёртка `require.New(t)` для удобства утверждений.

**Methods:**

#### `New(t *testing.T) (*Test, *require.Assertions)`

Создаёт и возвращает структуру `Test`, а также объект `require.Assertions`.

Автоматически выполняет:

- инициализацию конфигурации и логгера,
- генерацию короткого идентификатора теста,
- создание обёртки `require.New(t)`.

#### `(t *Test) Config() *config.Config`

Возвращает объект конфигурации, связанный с текущим тестом.

#### `(t *Test) Logger() log.Logger`

Возвращает логгер, связанный с текущим тестом.

#### `(t *Test) Assert() *require.Assertions`

Возвращает объект `require.Assertions` для выполнения проверок внутри теста.

#### `(t *Test) Id() string`

Возвращает идентификатор текущего теста (4 байта в hex-представлении).

#### `(t *Test) T() *testing.T`

Возвращает оригинальный объект `*testing.T`, переданный в `New()`.

## Usage

### Example usage in test

```go
package mypkg_test

import (
	"testing"
	"github.com/txix-open/isp-kit/test"
)

func TestSomething(t *testing.T) {
	testCtx, assert := test.New(t)

	cfg := testCtx.Config()
	logger := testCtx.Logger()
	id := testCtx.Id()

	logger.Info("Running test", log.String("id", id))
	assert.NotNil(cfg)
}
```

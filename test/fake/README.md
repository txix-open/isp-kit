# Package `fake`

Пакет `fake` предназначен для генерации тестовых данных произвольных структур с использованием библиотеки `go-faker/faker`.

## Types

### Option

Тип `Option` представляет собой алиас для `options.OptionFunc` из библиотеки `go-faker/faker`, используемый для настройки генерации случайных данных.

## Functions

### `It[T any](opts ...Option) T`

Генерирует случайное значение произвольного типа `T` с использованием указанных опций. По умолчанию устанавливает минимальный и максимальный размер срезов равным 1. Если генерация завершается с ошибкой, вызывает `panic`.

#### `MinSliceSize(value uint) Option`

Возвращает опцию, задающую минимальный размер случайно генерируемых срезов и отображений.

#### `MaxSliceSize(value uint) Option`

Возвращает опцию, задающую максимальный размер случайно генерируемых срезов и отображений.

## Usage

### Default usage flow

```go
package mypkg_test

import (
	"testing"
	"github.com/txix-open/isp-kit/test/fake"
	"github.com/stretchr/testify/require"
)

type User struct {
	Name  string
	Email string
	Tags  []string
}

func TestFakeUser(t *testing.T) {
	user := fake.It[User]()
	require.NotEmpty(t, user.Name)
	require.NotEmpty(t, user.Email)
	require.Len(t, user.Tags, 1) // благодаря Min/MaxSliceSize по умолчанию
}
```

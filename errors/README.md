# Package `errors`

Пакет `errors` предоставляет функции для создания, оборачивания и проверки ошибок, используя стандартные пакеты Go
`errors` и `fmt`.

Пакет заменяет привычные вызовы из `github.com/pkg/errors`.

## Functions

#### `New(text string) error`

Создает новую ошибку с переданным текстом.

#### `Errorf(format string, args ...any) error`

Создает новую ошибку с форматированием. Если формат содержит `%w`, ошибка будет обернута стандартным механизмом Go.

#### `WithMessage(err error, message string) error`

Добавляет сообщение к ошибке через стандартное оборачивание `%w`.

#### `WithMessagef(err error, format string, args ...any) error`

Добавляет форматированное сообщение к ошибке через стандартное оборачивание `%w`.

#### `Is(err, target error) bool`

Проверяет, содержит ли цепочка ошибок `target`.

#### `As(err error, target any) bool`

Находит в цепочке ошибок ошибку подходящего типа и записывает ее в `target`.

#### `AsType[E error](err error) (E, bool)`

Generic-версия `As`, возвращающая найденную ошибку и признак успешного поиска.

#### `Unwrap(err error) error`

Возвращает ошибку, обернутую внутри `err`.

#### `Join(errs ...error) error`

Объединяет несколько ошибок в одну.

## Usage

```go
package main

import (
	"database/sql"

	"github.com/txix-open/isp-kit/errors"
)

func getUser(userId string) error {
	err := sql.ErrNoRows
	if err != nil {
		return errors.WithMessagef(err, "get user %s", userId)
	}

	return nil
}
```

# Package `apierrors`

Пакет `apierrors` предоставляет инструменты для создания и управления структурированными ошибками HTTP-API.
Ошибки включают код, сообщение, детали и уровень логирования, а также автоматически сериализуются в JSON для ответов
клиентам.

## Types

### Error

Структура для представления ошибки API. Реализует интерфейс `error`.

**Methods:**

#### `NewInternalServiceError(err error) Error`

Создает ошибку для внутренних сбоев сервера. Возвращает ошибку с HTTP-статусом 500, бизнесовым кодом 900 и уровнем
логирования error.

#### `NewBusinessError(errorCode int, errorMessage string, err error) Error`

Создает бизнес-ошибку (например, невалидные данные от клиента). Возвращает ошибку с HTTP-статусом 400, указанным
бизнесовым кодом и уровнем логирования warn.

#### `New(httpStatusCode int, errorCode int, errorMessage string, err error) Error`

Создает кастомную ошибку с полным контролем параметров.

#### `(e Error) Error() string`

Получить строковое представление ошибки.

#### `(e Error) WriteError(w http.ResponseWriter) error`

Сериализует ошибку в JSON и записывает в `http.ResponseWriter`.

#### `(e Error) WithDetails(details map[string]any) Error`

Добавить детали информации об ошибке.

#### `(e Error) WithLogLevel(level log.Level) Error`

Изменить уровень логирования ошибки.

#### `(e Error) LogLevel() log.Level`

Получить текущий уровень логирования ошибки

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"errors"

	"github.com/txix-open/isp-kit/http/apierrors"
)

type User struct {
	Id   string
	Name string
}

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserNotFoundCode = 700
)

type userService interface {
	GetUserById(ctx context.Context, id string) (*User, error)
}

type userController struct {
	svc userService
}

func (c userController) GetUser(ctx context.Context, userId string) (*User, error) {
	result, err := c.svc.GetUserById(ctx, userId)
	switch {
	case errors.Is(err, ErrUserNotFound):
		/* user isn't found so we have to send back specific error code */
		return nil, apierrors.NewBusinessError(ErrUserNotFoundCode, ErrUserNotFound.Error(), err)
	case err != nil:
		/* got some internal error */
		return nil, apierrors.NewInternalServiceError(err)
	default:
		return result, nil
	}
}

```
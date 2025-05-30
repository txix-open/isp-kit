# Package `validator`

Пакет `validator` предоставляет адаптер для валидации структур с использованием библиотеки `github.com/txix-open/validator/v10` и поддержки перевода ошибок на английский язык.

## Types

### Adapter

Адаптер для валидации, инкапсулирующий валидатор и переводчик сообщений об ошибках.

**Methods:**

#### `New() Adapter`

Создаёт новый адаптер для валидатора с английским переводчиком. 

#### `Validate(v any) (ok bool, details map[string]string)`

Проверяет структуру, возвращает `true` если валидация прошла успешно, иначе `false` и карту ошибок по полям.

#### `ValidateToError(v any) error`

Проверяет структуру, возвращает `nil` если ошибок нет, иначе объект ошибки с подробным описанием.

## Usage

### Default usage flow

```go
package main

import (
	"fmt"
	"log"

	"github.com/txix-open/validator"
)

type User struct {
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"gte=0,lte=130"`
}

func main() {
	// Создаём адаптер валидатора с переводом ошибок на английский
	validator := validator.Default

	user := User{
		Name:  "",
		Email: "invalid-email",
		Age:   150,
	}

	ok, details := validator.Validate(user)
	if !ok {
		fmt.Println("Validation failed:")
		for field, errMsg := range details {
			fmt.Printf(" - %s: %s\n", field, errMsg)
		}
	} else {
		fmt.Println("Validation succeeded")
	}

	// Можно получить ошибку с описанием валидации
	err := validator.ValidateToError(user)
	if err != nil {
		log.Printf("Validation error: %v", err)
	}
}
```

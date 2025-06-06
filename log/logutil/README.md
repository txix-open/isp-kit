# Package `logutil`

Пакет `logutil` предоставляет функции-утилиты для гибкого управления уровнем логирования ошибок.

## Functions

#### `LogLevelForError(err error) log.Level`

Определить уровень логирования для ошибки.

#### `LogLevelFuncForError(err error, logger log.Logger) func(ctx context.Context, message any, fields ...log.Field)`

Получить функцию логирования соответствующего уровня для ошибки.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	log2 "github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/log/logutil"
)

type rateLimitError struct {
	Message string
}

func (e rateLimitError) Error() string {
	return e.Message
}

func (e rateLimitError) LogLevel() log2.Level {
	return log2.WarnLevel
}

// Обработка ошибки
func processRequest() error {
	/* put here logic with request */
	return rateLimitError{Message: "some error"}
}

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}

	err = processRequest()
	logFunc := logutil.LogLevelFuncForError(err, logger)
	logFunc(context.Background(),
		"Request processing failed",
		log2.String("error", err.Error()),
	)
}

```
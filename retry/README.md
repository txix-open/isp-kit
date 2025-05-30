# Package `retry`

Пакет `retry` предоставляет обёртку над `"github.com/cenkalti/backoff/v4"` для повторения операций с учётом времени и контекста.

## Types

### ExponentialBackoff

Структура `ExponentialBackoff` инкапсулирует в себе параметры экспоненциального бэкоффа, используемого для повторных попыток выполнения операций.

**Methods:**

#### `func NewExponentialBackoff(maxElapsedTime time.Duration) ExponentialBackoff`

Создаёт новый экземпляр `ExponentialBackoff` с заданным максимальным временем выполнения.

#### `(e ExponentialBackoff) Do(ctx context.Context, operation func() error) error`

Выполняет переданную операцию с использованием экспоненциального бэкоффа и поддержки контекста. Если операция завершается с ошибкой, она будет повторяться до истечения `maxElapsedTime` или отмены контекста.

## Usage

```go
package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/txix-open/isp-kit/retry"
)

func main() {
	retrier := retry.NewExponentialBackoff(10 * time.Second)

	err := retrier.Do(context.Background(), func() error {
		// Здесь должна быть логика, которую нужно повторить при ошибке
		log.Println("trying...")
		return errors.New("failed")
	})

	if err != nil {
		log.Fatalf("operation failed after retries: %v", err)
	}
}
```

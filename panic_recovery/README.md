# Package `panic_recovery`

Пакет `panic_recovery` предоставляет инструменты восстановления и обработки ошибки после паники.

## Functions

### Recover(formatRes func(err error))

Функция `Recover(formatRes func(err error))` выполняет восстановление после паники и вызывает функцию `formatRes()` 
последующей обработки ошибки.


## Usage

### Default usage flow

```go
package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/panic_recovery"
)

func main() {
	testPanic()
	fmt.Println("panic was recovered")
}

func testPanic() {
	defer panic_recovery.Recover(func(err error) {
		fmt.Printf("\ncatch panic err: %v\n", err)
	})

	panic(errors.New("fatal error"))
	return
}
```


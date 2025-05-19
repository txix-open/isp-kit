# Package `shutdown`

Пакет `shutdown` предназначен для перехвата системных сигналов завершения (SIGINT, SIGTERM) и выполнения заданной пользователем функции при их получении.

## Functions

### `func On(do func()) chan os.Signal`

Запускает обработку сигналов завершения процесса (SIGINT, SIGTERM) и вызывает переданную функцию `do` при получении одного из этих сигналов.

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/txix-open/isp-kit/shutdown"
)

func main() {
	shutdown.On(func() {
		fmt.Println("shutting down...")
	})

	fmt.Println("working...")
	time.Sleep(30 * time.Second)
}
```

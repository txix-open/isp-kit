# Package `lb`

Пакет `lb` предоставляет функциональность балансировки нагрузки.

## Types

### RoundRobin

Структура `RoundRobin` реализует алгоритм балансировки нагрузки Round Robin.

**Methods:**

#### `NewRoundRobin(hosts []string) *RoundRobin`

Конструктор, принимающий на вход список адресов для балансировки.

#### `(b *RoundRobin) Upgrade(hosts []string)`

Обновить список адресов.

#### `(b *RoundRobin) Size() int`

Получить текущее количество адресов.

#### `(b *RoundRobin) Next() (string, error)`

Получить следующий адрес по алгоритму Round Robin. Если список адресов пуст, то возвращается ошибка.

## Usage

### Default usage flow

```go
package main

import (
	"log"

	"github.com/txix-open/isp-kit/lb"
)

func main() {
	hostList := []string{"localhost:8080", "localhost:8081"}
	loadBalancer := lb.NewRoundRobin(hostList)

	/* host – random address from hostList*/
	host, err := loadBalancer.Next()
	if err != nil {
		log.Fatal(err)
	}
}

```
# Package `rc`

Пакет `rc` предоставляет механизм управления удалённой конфигурацией с возможностью её обновления, объединения с конфигурацией-переопределением и валидации.

## Types

### Validator

Интерфейс Validator определяет метод для валидации конфигурации.

### Config

Структура Config реализует управление конфигурацией с учётом переопределений, поддерживает блокировки и валидацию.

**Methods:**

#### `New(validator Validator, overrideData []byte) *Config`

Создаёт новый экземпляр Config с указанным валидатором и данными переопределения.

#### `(c *Config) Upgrade(data []byte, newConfigPtr any, prevConfigPtr any) error`

Обновляет конфигурацию:

* объединяет переданные данные с переопределениями,
* десериализует в `newConfigPtr`,
* валидирует конфигурацию,
* десериализует предыдущую конфигурацию в `prevConfigPtr`,
* сохраняет новую конфигурацию как предыдущую.

### Functions

#### `Upgrade[T any](rc *Config, data []byte) (newCfg T, prevCfg T, err error)`

Удобная обобщённая функция для обновления конфигурации с использованием типа T.

#### `GenerateConfigSchema(cfgPtr any) schema.Schema`

Генерирует схему конфигурации для переданной структуры с помощью генератора из пакета `rc/schema`.

## Usage
### Default usage flow

```go
import (
    "github.com/txix-open/isp-kit/rc"
)

type MyConfig struct {
    // поля конфигурации
}

type MyValidator struct{}

func (v MyValidator) ValidateToError(value any) error {
    // реализовать валидацию
    return nil
}

func main() {
    validator := MyValidator{}
    overrideData := []byte(`{"some.key":"override value"}`)
    rcConfig := rc.New(validator, overrideData)

    data := []byte(`{"some.key":"value"}`)

    var newCfg, prevCfg MyConfig
    err := rcConfig.Upgrade(data, &newCfg, &prevCfg)
    if err != nil {
        panic(err)
    }
    // newCfg - новая конфигурация
    // prevCfg - предыдущая конфигурация
}
```
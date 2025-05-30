# Package `rct`

Пакет `rct` предназначен для тестирования схемы и валидации удалённой конфигурации, используемой с `isp-kit`.

## Functions

### `Test[T any](t *testing.T, defaultRemoteConfigPath string, remoteConfig T)`

Проверяет, что удалённая конфигурация соответствует сгенерированной JSON-схеме и проходит валидацию.

* Проверяет отсутствие устаревшего тега `valid`
* Проверяет соответствие структуры JSON-схеме
* Выполняет валидацию значений с помощью `validator.Default`

### `FindTag[T any](v T, tag string) bool`

Рекурсивно проверяет, содержит ли структура тег с указанным именем.

## Usage

### Example usage in test

```go
package mypkg_test

import (
	"testing"

	"github.com/txix-open/isp-kit/rct"
)

type RemoteCfg struct {
	Name string `json:"name" validate:"required"`
}

func TestRemoteCfg(t *testing.T) {
	rct.Test(t, "conf/default_remote_config.json", RemoteCfg{})
}
```

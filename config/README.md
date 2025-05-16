# Package `config`

Пакет `config` предназначен для работы с конфигурацией приложения с поддержкой:

- множественных источников данных (yaml, переменные окружения, кастомные источники)
- валидации конфигурации
- типизированного доступа к параметрам

## Types

### Config

Основная структура для работы с конфигурацией.

**Methods:**

#### `New(opts ...Option) (*Config, error)`

Конструктор конфигурации с опциями:

- `WithExtraSource` - добавить дополнительные источники, реализующие интерфейс `Source`
- `WithEnvPrefix` - установить префикс для переменных окружения
- `WithValidator` - установить валидатор конфигурации, реализующий интерфейс `Validator`

#### `(c *Config) Set(key string, value any)`

Установить новое/перезаписать существующее значение параметра конфигурации.

#### `(c *Config) Delete(key string)`

Удалить параметра конфигурации.

#### `(c *Config) Mandatory() Mandatory`

Получить обязательные параметры конфигурации.

#### `(c *Config) Optional() Optional`

Получить опциональные параметры конфигурации (возвращает значения по-умолчанию).

#### `(c *Config) Read(ptr any) error`

Спарсить всю конфигурацию и записать ее по переданному указателю. Если был установлен валидатор, то спаршенная конфигурация будет провалидирована.

### YamlFileSource

Создание источника конфигурации из yaml-файла.

#### `(y YamlFileSource) Config() (map[string]string, error)`
Получить конфигурацию из yaml-файла в виде мэпы.

## Usage

### Default usage flow

```go
package main

import (
	"log"
	"time"

	"github.com/txix-open/isp-kit/config"
	"github.com/txix-open/isp-kit/validator"
)

type appConfig struct {
	Host     string `validate:"required"`
	Port     int    `validate:"required"`
	Timeout  time.Duration
	Database struct {
		Dsn string
	}
}

func main() {
	cfg, err := config.New(
		config.WithExtraSource(config.NewYamlConfig("config.yml")),
		config.WithEnvPrefix("APP_"),
		config.WithValidator(validator.Default),
	)
	if err != nil {
		log.Fatal(err)
	}

	var actual appConfig
	err = cfg.Read(&actual)
	if err != nil {
		log.Fatal(err)
	}

	/* access to parameters */
	port := cfg.Mandatory().Int("port")
	timeout := cfg.Optional().Duration("timeout", 5*time.Second)
}

```
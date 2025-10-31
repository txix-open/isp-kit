# Package `bootstrap`

Пакет `bootstrap` предназначен для инициализации инфраструктуры приложения и настройки основных компонентов.

## Types

### Bootstrap

Центральная структура, которая предоставляет функциональность для инициализации и настройки модуля приложения, включая
динамическую конфигурацию, логирование, трейсинг, запуск инфраструктурного http-сервера (метрики, healthchecks),
подключение к кластеру сервисов и интеграцию с sentry для обработки ошибок. Другими словами, объект структуры
`Bootstrap` управляет жизненным циклом приложения.

**Методы:**

#### `(b *BaseBootstrap) Fatal(err error)`

Обработать критические ошибки с уведомлением в Sentry

#### `New(moduleVersion string, remoteConfig any, endpoints []cluster.EndpointDescriptor) *Bootstrap`

Конструктор с параметрами:

- `moduleVersion` - версия модуля
- `remoteConfig` - структура для динамической конфигурации
- `endpoints` - список эндпоинтов модуля

#### `NewStandalone(moduleVersion string) *StandaloneBootstrap`

Конструктор с параметрами:

- `moduleVersion` - версия модуля

#### `func (b *StandaloneBootstrap) ReadConfig(destPtr any) error`

Прочитать локальную `json` конфигурацию, путь до конфигурации определяется настройкой в локальном конфиге `remoteConfigPath`, по умолчанию путь `./conf/config.json` (либо относительно бинарника, если запуск не в `dev` режиме)

## Конфигурация

Файлы конфигурации по умолчанию:

- `conf/config_dev.yml` для разработки
- `config.yml` рядом с бинарником для `production`

Переменные окружения:

- `APP_MODE=dev` — режим разработки
- `APP_CONFIG_PATH` — кастомный путь к конфигу
- `APP_CONFIG_ENV_PREFIX` — префикс для env variables
- `CLUSTER_MODE=offline` — режим, при котором будет использоваться заглушка для конфиг сервиса

## Инфраструктурные эндпоинты

По умолчанию доступны:

- `/internal/metrics` — prometheus метрики
- `/internal/metrics/descriptions` — описание метрик
- `/internal/health` — healthcheck статус
- `/internal/debug/pprof/` — профилирование

## Usage

### Default usage flow for clustered

```go
package main

import (
	"log"

	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/shutdown"
)

type remoteConfig struct {
	Foo string `validate:"required"`
	Bar int
}

func noopHandler() {}

func main() {
	endpoints := []cluster.EndpointDescriptor{{
		Path:    "/api/service/private",
		Inner:   true,
		Handler: noopHandler,
	}, {
		Path:    "/api/service/public",
		Inner:   false,
		Handler: noopHandler,
	}}
	boot := bootstrap.New("1.0.0", remoteConfig{}, endpoints)

	shutdown.On(func() { /* waiting for SIGINT & SIGTERM signals */
		log.Println("shutting down...")
		boot.App.Shutdown()
		log.Println("shutdown completed")
	})

	err := boot.App.Run()
	if err != nil {
		boot.Fatal(err)
	}
}

```

### Default usage flow for standalone

```go
package main

import (
	"log"

	"github.com/txix-open/isp-kit/bootstrap"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/shutdown"
)

type config struct {
	Foo string `validate:"required"`
	Bar int
}

func main() {
	boot := bootstrap.NewStandalone("1.0.0")

	shutdown.On(func() { /* waiting for SIGINT & SIGTERM signals */
		log.Println("shutting down...")
		boot.App.Shutdown()
		log.Println("shutdown completed")
	})

	cfg := config{}
	err := boot.ReadConfig(&cfg)
	if err != nil {
		boot.Fatal(err)
	}

	err = boot.App.Run()
	if err != nil {
		boot.Fatal(err)
	}
}

```

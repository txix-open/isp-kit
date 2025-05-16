# Package `app`

Пакет `app` предназначен для создания приложений с простым и гибким управлением жизненным циклом.

## Types

### Application

Структура `Application` представляет собой абстракцию, которая содержит в себе основные необходимые компоненты
приложения (logger, config, context), а также позволяет простым и гибким способом управлять жизненным циклом компонент
приложения реализующих интерфейсы `Runner` и `Closer`.

**Methods:**

#### `(a *Application) Context() context.Context`

Получить контекст приложения.

#### `(a *Application) Config() *config.Config`

Получить конфигурацию приложения.

#### `(a *Application) Logger() *log.Adapter`

Получить логгер.

#### `(a *Application) AddRunners(runners ...Runner)`

Добавить компоненты приложения, реализующих интерфейс `Runner`.

#### `(a *Application) AddClosers(closers ...Closer)`

Добавить компоненты приложения, реализующих интерфейс `Closer`.

#### `(a *Application) Run() error`

Вызывает у каждой `Runner` компоненты приложения метод `Run`. Блокирующая операция. Если хотя бы одна компонента при
запуске вернула ошибку, то метод `Run` завершается с ошибкой.

#### `(a *Application) Close()`

Вызывает у каждой `Closer` компоненты приложения метод `Close`.

#### `(a *Application) Shutdown()`

То же самое, что и метод `Close`, но в конце дополнительно завершает контекст.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/shutdown"
)

func noopRunnerFn(_ context.Context) error { return nil }

func noopCloserFn() error { return nil }

type noopRunner struct{}

func (r noopRunner) Run(_ context.Context) error { return nil }

type noopCloser struct{}

func (c noopCloser) Close() error { return nil }

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
	application.AddRunners(app.RunnerFunc(noopRunnerFn), noopRunner{})
	application.AddClosers(app.CloserFunc(noopCloserFn), noopCloser{})

	shutdown.On(func() { /* waiting for SIGINT & SIGTERM signals */
		log.Println("shutting down...")
		application.Shutdown()
		log.Println("shutdown completed")
	})

	err = application.Run() /* blocking here */
	if err != nil {
		log.Fatal(err)
	}
}

```

### Construct `Application` instance with options

```go
package main

import (
	"log"

	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/config"
	log2 "github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/validator"
)

func main() {
	application, err := app.New(
		app.WithConfigOptions(
			config.WithValidator(validator.Default),
			/* pass the config options you want */
		),
		app.WithLoggerConfigSupplier(func(cfg *config.Config) log2.Config {
			var loggerCfg log2.Config
			/* form the logger config the way you want */
			return loggerCfg
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	...
}

```

### Construct `Application` instance with config

```go
package main

import (
	"log"

	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/config"
	"github.com/txix-open/isp-kit/validator"
)

func main() {
	config := app.Config{
		ConfigOptions: []config.Option{
			config.WithValidator(validator.Default),
			/* pass the config options you want */
		},
	}
	application, err := app.NewFromConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	...
}

```
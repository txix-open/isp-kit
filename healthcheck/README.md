# Package `healthcheck`

Пакет `healthcheck` предоставляет инструменты для реализации и управления проверками состояния (health checks) в
приложениях. Он позволяет регистрировать проверки компонентов, кэшировать результаты и предоставлять их через
HTTP-эндпоинт в формате JSON.

## Types

### Registry

Структура `Registry` управляет проверками состояния и кэширует их результаты.

**Methods:**

#### `NewRegistry() *Registry`

Конструктор реестра проверок состояния.

#### `(r *Registry) Register(name string, checker Checker)`

Зарегистрировать компоненту проверки состояния с именем `name`. Такой компонентой может быть объект, реализующий
интерфейс `Checker`, либо же функция, обернутая в тип `CheckerFunc`.

#### `(r *Registry) Handler() http.Handler`

HTTP-обработчик для интеграции с HTTP-сервером.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"
	"net/http"

	"github.com/txix-open/isp-kit/dbrx"
	"github.com/txix-open/isp-kit/grmqx"
	"github.com/txix-open/isp-kit/healthcheck"
	log2 "github.com/txix-open/isp-kit/log"
)

func main() {
	logger, err := log2.New()
	if err != nil {
		log.Fatal(err)
	}
	var (
		dbCli  = dbrx.New(logger)
		rmqCli = grmqx.New(logger)
	)

	registry := healthcheck.NewRegistry()
	registry.Register("database", dbCli)
	registry.Register("mq", rmqCli)
	registry.Register("auth-service", healthcheck.CheckerFunc(func(ctx context.Context) error {
		/* health check auth service */
		return nil
	}))

	/* integration with HTTP-server */
	http.Handle("/health", registry.Handler())
}

```
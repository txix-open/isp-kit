# Package `cluster`

Пакет `cluster` предназначен для управления кластером микросервисов с возможностями:

- Регистрация модулей в кластере
- Динамическое обновление конфигурации
- Обмен маршрутами между сервисами
- Обработка событий кластера

## Types

### Client

Структура `Client` представляет собой клиент кластера с функциональностью установки и поддержки соединения с
конфигурационным сервисом, регистрации модуля в кластере, обработки жизненного цикла соединения с отправкой метаданных
модуля (версия, эндпоинты)

**Methods:**

#### `NewClient(moduleInfo ModuleInfo, configData ConfigData, hosts []string, logger log.Logger) *Client`

Конструктор клиента кластера.

#### `(c *Client) Run(ctx context.Context, eventHandler *EventHandler) error`

Запустить клиент кластера с указанным обработчиком событий.

#### `(c *Client) Close() error`

Завершить работу клиента.

#### `(c *Client) Healthcheck(ctx context.Context) error`

Проверить активна ли текущая сессия.

### EventHandler

Структура `EventHandler` представляет собой компоновку обработчиков основных событий в кластере микросервисов.

**Methods:**

#### `NewEventHandler() *EventHandler`

Конструктор.

#### `(h *EventHandler) RemoteConfigReceiverWithTimeout(receiver RemoteConfigReceiver, timeout time.Duration) *EventHandler`

Установить обработчик на получение обновленной динамической-конфигурации текущего модуля. Обработчик должен реализовывать интерфейс `RemoteConfigReceiver`.

#### `(h *EventHandler) RemoteConfigReceiver(receiver RemoteConfigReceiver) *EventHandler`

То же самое, что и `RemoteConfigReceiverWithTimeout`, но без возможности указать timeout на обработку конфига. Используется значение по умолчанию в 5 секунд.

#### `(h *EventHandler) RoutesReceiver(receiver RoutesReceiver) *EventHandler`

Установить обработчик на получение актуальных эндпоинтов, зависимостях и прочей информации о сервисах в кластере. Обработчик должен реализовывать интерфейс `RoutesReceiver`.

#### `(h *EventHandler) RequireModule(moduleName string, upgrader HostsUpgrader) *EventHandler`

Установить обработчик на получение списка адресов модулей, от которых зависит текущий сервис. Обработчик должен реализовывать интерфейс `HostsUpgrader`.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"log"

	"github.com/txix-open/isp-kit/app"
	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/rc"
	"github.com/txix-open/isp-kit/shutdown"
)

type remoteConfigUpdater struct{}

func (r remoteConfigUpdater) ReceiveConfig(ctx context.Context, remoteConfig []byte) error {
	/* handle & update new remote config */
	return nil
}

type routesHandler struct{}

func (r routesHandler) ReceiveRoutes(ctx context.Context, routes cluster.RoutingConfig) error {
	for _, module := range routes {
		/* handle each module's info */
	}
	return nil
}

func noopHandler() {}

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	moduleInfo := cluster.ModuleInfo{
		ModuleName: "noop-service",
		Version:    "1.0.0",
		Endpoints: []cluster.EndpointDescriptor{{
			Path:    "/api/foo",
			Inner:   false,
			Handler: noopHandler,
		}},
	}
	defaultRemoteConfig := []byte{...}
	configData := cluster.ConfigData{
		Version: "1.0.0",
		Schema:  rc.GenerateConfigSchema(defaultRemoteConfig),
		Config:  defaultRemoteConfig,
	}
	clusterCli := cluster.NewClient(
		moduleInfo,
		configData,
		[]string{"config-service:8080"},
		application.Logger(),
	)

	/* client implements HostsUpgrader interface & will have auth-service's addresses */
	authServiceCli, err := client.Default()
	if err != nil {
		log.Fatal(err)
	}

	eventHandler := cluster.NewEventHandler(). /* Настройка обработчиков */
		RemoteConfigReceiver(remoteConfigUpdater{}).
		RoutesReceiver(routesHandler{}).
		RequireModule("auth-service", authServiceCli)
	application.AddRunners(app.RunnerFunc(func(ctx context.Context) error {
		err := clusterCli.Run(ctx, eventHandler)
		if err != nil {
			return err
		}
		return nil
	}))
	application.AddClosers(clusterCli, authServiceCli)

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
# Package `common_endpoints`

Пакет `common_endpoints` предназначен для формирования набора стандартных (общих) эндпоинтов сервиса.

## Types

### CommonEndpointOption

Тип `CommonEndpointOption` представляет собой функциональную опцию для конфигурации набора общих эндпоинтов.

Используется для модификации внутренней конфигурации `commonEndpointsCfg`.

```go
type CommonEndpointOption func(cfg *commonEndpointsCfg)
```

## Functions

### `CommonEndpoints(basePath string, opts ...CommonEndpointOption) []cluster.EndpointDescriptor`

Создаёт и возвращает список `cluster.EndpointDescriptor` на основе переданных опций.

* `basePath` — базовый путь для всех создаваемых эндпоинтов.
* `opts` — набор опций конфигурации (например, добавление swagger-endpoint).

### `WithSwaggerEndpoint(swagger []byte) CommonEndpointOption`

Добавляет endpoint для получения Swagger-спецификации.

* `swagger` — содержимое swagger-файла в виде `[]byte`, если он пустой, endpoint не будет добавлен.
* Данные автоматически кодируются в Base64 перед возвратом клиенту.

Endpoint будет доступен по пути:

```
{basePath}/swagger
```

и будет иметь параметры:

* HTTP метод: `GET`
* `UserAuthRequired = false`
* `Inner = false`


## Internal behavior

### Swagger endpoint

Swagger endpoint формируется функцией `swaggerEndpoint` и возвращает `cluster.EndpointDescriptor` со следующими характеристиками:

* Path: `basePath + "/swagger"`
* Handler: возвращает Base64-кодированную строку swagger-файла
* Не требует аутентификации пользователя
* Использует метод `GET`

## Usage

### Default usage flow

```go
package main

import (
	"os"

	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/common_endpoints"
)

func main() {
	swaggerBytes, _ := os.ReadFile("swagger.json")

	endpoints := common_endpoints.CommonEndpoints(
		"/api",
		common_endpoints.WithSwaggerEndpoint(swaggerBytes),
	)

	var descriptors []cluster.EndpointDescriptor
	descriptors = append(descriptors, endpoints...)

	// передаём descriptors в config сервис/router
}
```

### Without swagger endpoint

```go
endpoints := common_endpoints.CommonEndpoints("/api")
// endpoints будет пустым
```
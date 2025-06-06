# Package `client`

Пакет `client` предоставляет SOAP-клиент для взаимодействия с SOAP-сервисами. Поддерживает отправку запросов, обработку
ответов и ошибок в формате SOAP Fault.

## Types

### Client

Структура для отправки SOAP-запросов. Использует [`httpcli.Client`](../../httpcli/client.go) для HTTP-взаимодействия.

**Methods:**

#### `New(cli *httpcli.Client) Client`

Конструктор клиента.

####

`(c Client) Invoke(ctx context.Context, url string, soapAction string, extraHeaders map[string]string, requestBody any) (*Response, error)`

Отправляет SOAP-запрос. Автоматически формирует SOAP-конверт.

## Usage

### Default usage flow

```go
package main

import (
	"context"
	"encoding/xml"
	"log"

	"github.com/txix-open/isp-kit/http/httpclix"
	"github.com/txix-open/isp-kit/http/soap/client"
)

type userRequest struct {
	XMLName xml.Name `xml:"UserRequest"`
	Name    string   `xml:"name"`
}

type userResponse struct {
	XMLName xml.Name `xml:"UserResponse"`
	Id      int      `xml:"id"`
}

func main() {
	cli := client.New(httpclix.Default())
	resp, err := cli.Invoke(
		context.Background(),
		"https://api.example.com/users",
		"CreateUser",
		nil,
		userRequest{Name: "Alice"},
	)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Close()

	var res userResponse
	err = resp.UnmarshalPayload(&res)
	if err == nil {
		log.Printf("id: %d\n", res.Id)
		return
	}

	fault, err := resp.Fault()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("SOAP Fault: %s\n", fault.String)
}

```
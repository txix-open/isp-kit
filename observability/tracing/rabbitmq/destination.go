package rabbitmq

import (
	"fmt"
)

func Destination(exchange string, routingKey string) string {
	if exchange == "" {
		return routingKey
	}
	return fmt.Sprintf("%s/%s", exchange, routingKey)
}

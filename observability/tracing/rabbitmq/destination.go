package rabbitmq

import (
	"fmt"
)

// Destination creates a destination identifier from an exchange and routing key.
// If the exchange is empty, it returns just the routing key; otherwise, it returns
// the exchange and routing key joined by a forward slash.
func Destination(exchange string, routingKey string) string {
	if exchange == "" {
		return routingKey
	}
	return fmt.Sprintf("%s/%s", exchange, routingKey)
}

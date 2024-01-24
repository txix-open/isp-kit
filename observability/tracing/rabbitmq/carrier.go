package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type TableCarrier amqp091.Table

func (c TableCarrier) Get(key string) string {
	value := c[key]
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

func (c TableCarrier) Set(key string, value string) {
	c[key] = value
}

func (c TableCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for key := range c {
		keys = append(keys, key)
	}
	return keys
}

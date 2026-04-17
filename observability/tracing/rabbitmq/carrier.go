// Package rabbitmq provides utilities for RabbitMQ tracing integration.
package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// TableCarrier implements the TextMapCarrier interface for AMQP message headers.
type TableCarrier amqp091.Table

// Get returns the string representation of the value associated with the given key.
// It returns an empty string if the key does not exist.
func (c TableCarrier) Get(key string) string {
	value := c[key]
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// Set associates a key-value pair in the message headers.
func (c TableCarrier) Set(key string, value string) {
	c[key] = value
}

// Keys returns all keys in the message headers.
func (c TableCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for key := range c {
		keys = append(keys, key)
	}
	return keys
}

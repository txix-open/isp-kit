// Package stompt provides test helpers for STOMP messaging operations
// (typically used with ActiveMQ). It creates isolated queues for each test.
package stompt

import (
	"context"
	"fmt"

	"github.com/txix-open/isp-kit/stompx"
	"github.com/txix-open/isp-kit/test"
)

// Client provides a test helper for STOMP messaging operations.
// It manages isolated queues (prefixed with test ID) for each test.
type Client struct {
	t        *test.Test
	address  string
	username string
	password string
}

// New creates a new STOMP test client.
// Connection parameters can be overridden using environment variables:
// ACTIVEMQ_STOMP_ADDRESS, ACTIVEMQ_USERNAME, ACTIVEMQ_PASSWORD.
func New(t *test.Test) *Client {
	return &Client{
		t:        t,
		address:  t.Config().Optional().String("ACTIVEMQ_STOMP_ADDRESS", "127.0.0.1:61613"),
		username: t.Config().Optional().String("ACTIVEMQ_USERNAME", "test"),
		password: t.Config().Optional().String("ACTIVEMQ_PASSWORD", "test"),
	}
}

// ConsumerConfig returns a ConsumerConfig for the specified queue,
// with the queue name prefixed by the test ID for isolation.
func (c *Client) ConsumerConfig(queue string) stompx.ConsumerConfig {
	return stompx.ConsumerConfig{
		Address:       c.address,
		Queue:         fmt.Sprintf("%s_%s", c.t.Id(), queue),
		Concurrency:   1,
		PrefetchCount: 1,
		Username:      c.username,
		Password:      c.password,
		ConnHeaders:   nil,
	}
}

// PublisherConfig returns a PublisherConfig for the specified queue,
// with the queue name prefixed by the test ID for isolation.
func (c *Client) PublisherConfig(queue string) stompx.PublisherConfig {
	return stompx.PublisherConfig{
		Address:     c.address,
		Queue:       fmt.Sprintf("%s_%s", c.t.Id(), queue),
		Username:    c.username,
		Password:    c.password,
		ConnHeaders: nil,
	}
}

// Upgrade creates a new STOMP group and upgrades it with the provided
// configuration. The group is automatically closed when the test completes.
// Panics if upgrade fails.
func (c *Client) Upgrade(config stompx.Config) {
	group := stompx.New(c.t.Logger())
	c.t.T().Cleanup(func() {
		err := group.Close()
		c.t.Assert().NoError(err)
	})
	err := group.Upgrade(context.Background(), config)
	c.t.Assert().NoError(err)
}

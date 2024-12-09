package stompt

import (
	"context"
	"fmt"

	"github.com/txix-open/isp-kit/stompx"
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/test"
)

type Client struct {
	t        *test.Test
	address  string
	username string
	password string
}

func New(t *test.Test) *Client {
	return &Client{
		t:        t,
		address:  t.Config().Optional().String("ACTIVEMQ_STOMP_ADDRESS", "127.0.0.1:61613"),
		username: t.Config().Optional().String("ACTIVEMQ_USERNAME", "test"),
		password: t.Config().Optional().String("ACTIVEMQ_PASSWORD", "test"),
	}
}

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

func (c *Client) PublisherConfig(queue string) stompx.PublisherConfig {
	return stompx.PublisherConfig{
		Address:     c.address,
		Queue:       fmt.Sprintf("%s_%s", c.t.Id(), queue),
		Username:    c.username,
		Password:    c.password,
		ConnHeaders: nil,
	}
}

func (c *Client) Upgrade(consumers ...consumer.Config) {
	group := stompx.NewConsumerGroup(c.t.Logger())
	c.t.T().Cleanup(func() {
		err := group.Close()
		c.t.Assert().NoError(err)
	})
	err := group.Upgrade(context.Background(), consumers...)
	c.t.Assert().NoError(err)
}

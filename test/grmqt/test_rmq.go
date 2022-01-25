package grmqt

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/integration-system/isp-kit/grmqx"
	"github.com/integration-system/isp-kit/test"
	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	cfg  grmqx.Connection
	t    *test.Test
	conn *amqp091.Connection
}

func New(t *test.Test) *Client {
	host := t.Config().Optional().String("RMQ_HOST", "127.0.0.1")
	port := t.Config().Optional().Int("RMQ_PORT", 5672)
	user := t.Config().Optional().String("RMQ_USER", "guest")
	pass := t.Config().Optional().String("RMQ_PASS", "guest")
	vhost := fmt.Sprintf("test_%s_%s", t.Id(), strings.ToLower(t.T().Name()))

	vhostUrl := fmt.Sprintf("http://%s:15672/api/vhosts/%s", host, vhost)

	req, err := http.NewRequest(http.MethodPut, vhostUrl, nil)
	t.Assert().NoError(err)
	req.SetBasicAuth(user, pass)
	resp, err := http.DefaultClient.Do(req)
	t.Assert().NoError(err)
	t.Assert().EqualValues(201, resp.StatusCode)

	t.T().Cleanup(func() {
		req, err := http.NewRequest(http.MethodDelete, vhostUrl, nil)
		t.Assert().NoError(err)
		req.SetBasicAuth(user, pass)
		resp, err := http.DefaultClient.Do(req)
		t.Assert().NoError(err)
		t.Assert().EqualValues(204, resp.StatusCode)
	})

	cfg := grmqx.Connection{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		Vhost:    vhost,
	}

	conn, err := amqp091.Dial(cfg.Url())
	t.Assert().NoError(err)
	t.T().Cleanup(func() {
		err := conn.Close()
		t.Assert().NoError(err)
	})

	return &Client{
		cfg:  cfg,
		t:    t,
		conn: conn,
	}
}

func (c *Client) ConnectionConfig() grmqx.Connection {
	return c.cfg
}

func (c *Client) QueueLength(queue string) int {
	ch, err := c.conn.Channel()
	c.t.Assert().NoError(err)
	defer func() {
		err := ch.Close()
		c.t.Assert().NoError(err)
	}()

	q, err := ch.QueueInspect(queue)
	c.t.Assert().NoError(err)
	return q.Messages
}

func (c *Client) Publish(exchange string, routingKey string, messages ...amqp091.Publishing) {
	ch, err := c.conn.Channel()
	c.t.Assert().NoError(err)
	defer func() {
		err := ch.Close()
		c.t.Assert().NoError(err)
	}()

	for _, message := range messages {
		err := ch.Publish(exchange, routingKey, true, false, message)
		c.t.Assert().NoError(err)
	}
}

func (c *Client) DrainMessage(queue string) amqp091.Delivery {
	ch, err := c.conn.Channel()
	c.t.Assert().NoError(err)
	defer func() {
		err := ch.Close()
		c.t.Assert().NoError(err)
	}()

	msg, got, err := ch.Get(queue, true)
	c.t.Assert().NoError(err)

	c.t.Assert().True(got, "at least 1 message is expected")

	return msg
}

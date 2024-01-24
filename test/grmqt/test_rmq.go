package grmqt

import (
	"context"
	"fmt"
	"net/http"

	"github.com/integration-system/isp-kit/grmqx"
	"github.com/integration-system/isp-kit/json"
	"github.com/integration-system/isp-kit/test"
	"github.com/rabbitmq/amqp091-go"
)

type Client struct {
	connCfg  grmqx.Connection
	t        *test.Test
	conn     *amqp091.Connection
	GrmqxCli *grmqx.Client
}

func New(t *test.Test) *Client {
	host := t.Config().Optional().String("RMQ_HOST", "127.0.0.1")
	port := t.Config().Optional().Int("RMQ_PORT", 5672)
	user := t.Config().Optional().String("RMQ_USER", "guest")
	pass := t.Config().Optional().String("RMQ_PASS", "guest")
	vhost := fmt.Sprintf("test_%s", t.Id())

	vhostUrl := fmt.Sprintf("http://%s:15672/api/vhosts/%s", host, vhost)

	req, err := http.NewRequest(http.MethodPut, vhostUrl, nil)
	t.Assert().NoError(err)
	req.SetBasicAuth(user, pass)
	resp, err := http.DefaultClient.Do(req)
	t.Assert().NoError(err)
	t.Assert().EqualValues(http.StatusCreated, resp.StatusCode)

	t.T().Cleanup(func() {
		req, err := http.NewRequest(http.MethodDelete, vhostUrl, nil)
		t.Assert().NoError(err)
		req.SetBasicAuth(user, pass)
		resp, err := http.DefaultClient.Do(req)
		t.Assert().NoError(err)
		t.Assert().EqualValues(http.StatusNoContent, resp.StatusCode)
	})

	connCfg := grmqx.Connection{
		Host:     host,
		Port:     port,
		Username: user,
		Password: pass,
		Vhost:    vhost,
	}

	conn, err := amqp091.Dial(connCfg.Url())
	t.Assert().NoError(err)
	t.T().Cleanup(func() {
		err := conn.Close()
		t.Assert().NoError(err)
	})

	grmqxCli := grmqx.New(t.Logger())
	t.T().Cleanup(func() {
		grmqxCli.Close()
	})

	return &Client{
		connCfg:  connCfg,
		t:        t,
		conn:     conn,
		GrmqxCli: grmqxCli,
	}
}

func (c *Client) ConnectionConfig() grmqx.Connection {
	return c.connCfg
}

func (c *Client) QueueLength(queue string) int {
	var (
		q   amqp091.Queue
		err error
	)
	c.useChannel(func(ch *amqp091.Channel) {
		q, err = ch.QueueInspect(queue) //nolint
		c.t.Assert().NoError(err)
	})
	return q.Messages
}

func (c *Client) Upgrade(config grmqx.Config) {
	config.Url = c.connCfg.Url()
	err := c.GrmqxCli.Upgrade(context.Background(), config)
	c.t.Assert().NoError(err)
}

func (c *Client) PublishJson(exchange string, routingKey string, data any) {
	body, err := json.Marshal(data)
	c.t.Assert().NoError(err)
	pub := amqp091.Publishing{
		Body:        body,
		ContentType: "application/json",
	}
	c.Publish(exchange, routingKey, pub)
}

func (c *Client) Publish(exchange string, routingKey string, messages ...amqp091.Publishing) {
	c.useChannel(func(ch *amqp091.Channel) {
		for _, message := range messages {
			err := ch.PublishWithContext(context.Background(), exchange, routingKey, true, false, message)
			c.t.Assert().NoError(err)
		}
	})
}

func (c *Client) DrainMessage(queue string) amqp091.Delivery {
	var (
		msg amqp091.Delivery
		got bool
		err error
	)
	c.useChannel(func(ch *amqp091.Channel) {
		msg, got, err = ch.Get(queue, true)
		c.t.Assert().NoError(err)
	})

	c.t.Assert().True(got, "at least 1 message is expected")

	return msg
}

func (c *Client) useChannel(f func(ch *amqp091.Channel)) {
	ch, err := c.conn.Channel()
	c.t.Assert().NoError(err)
	defer func() {
		err := ch.Close()
		c.t.Assert().NoError(err)
	}()
	f(ch)
}

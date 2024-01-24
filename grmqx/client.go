package grmqx

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/integration-system/grmq"
	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
)

const (
	DefaultHeartbeat   = 3 * time.Second
	DefaultDialTimeout = 5 * time.Second
)

type Client struct {
	logger  log.Logger
	cli     *grmq.Client
	prevCfg Config
	lock    sync.Locker
}

func New(logger log.Logger) *Client {
	return &Client{
		logger:  logger,
		lock:    &sync.Mutex{},
		cli:     nil,
		prevCfg: Config{},
	}
}

func (c *Client) Upgrade(ctx context.Context, config Config) error {
	return c.upgrade(ctx, config, false)
}

func (c *Client) UpgradeAndServe(ctx context.Context, config Config) {
	_ = c.upgrade(ctx, config, true)
}

func (c *Client) Healthcheck(ctx context.Context) error {
	if c.prevCfg.Url == "" {
		return errors.New("client is not initialized")
	}
	cli := grmq.New(c.prevCfg.Url)
	err := cli.Run(ctx)
	if err != nil {
		return errors.WithMessage(err, "connect to rabbit mq")
	}
	cli.Shutdown()
	return nil
}

func (c *Client) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.prevCfg = Config{}
	cli := c.cli
	c.cli = nil
	if cli != nil {
		cli.Shutdown()
	}
}

func (c *Client) upgrade(ctx context.Context, config Config, justServe bool) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.logger.Debug(ctx, "rmq client: received new config")

	if reflect.DeepEqual(c.prevCfg, config) {
		c.logger.Debug(ctx, "rmq client: configs are equal. skipping initialization")
		return nil
	}

	c.logger.Debug(ctx, "rmq client: initialization began")

	if c.cli != nil {
		c.cli.Shutdown()
		c.cli = nil
	}

	cli := grmq.New(
		config.Url,
		grmq.WithDialConfig(grmq.DialConfig{
			Config: amqp091.Config{
				Heartbeat: DefaultHeartbeat,
				Locale:    "en_US",
			},
			DialTimeout: DefaultDialTimeout,
		}),
		grmq.WithPublishers(config.Publishers...),
		grmq.WithConsumers(config.Consumers...),
		grmq.WithDeclarations(config.Declarations),
		grmq.WithObserver(NewLogObserver(ctx, c.logger)),
	)

	if justServe {
		cli.Serve(ctx)
	} else {
		err := cli.Run(ctx)
		if err != nil {
			return err
		}
	}

	c.cli = cli
	c.prevCfg = config
	return nil
}

package grmqx

import (
	"context"
	"reflect"
	"sync"

	"github.com/integration-system/grmq"
	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
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
		grmq.WithPublishers(config.Publishers...),
		grmq.WithConsumers(config.Consumers...),
		grmq.WithDeclarations(config.Declarations),
		grmq.WithObserver(NewLogObserver(ctx, c.logger)),
	)
	err := cli.Run(ctx)
	if err != nil {
		return err
	}

	c.cli = cli
	c.prevCfg = config

	return nil

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

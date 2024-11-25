package kafkax

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
	"golang.org/x/sync/errgroup"
)

type Client struct {
	logger  log.Logger
	prevCfg Config
	cli     *KafkaClient
	lock    sync.Locker
}

func New(logger log.Logger) *Client {
	return &Client{
		logger:  logger,
		prevCfg: Config{},
		cli:     nil,
		lock:    &sync.Mutex{},
	}
}

type KafkaClient struct {
	publishers []*publisher.Publisher
	consumers  []consumer.Consumer
	observer   Observer
}

func (c *Client) UpgradeAndServe(ctx context.Context, config Config) error {
	return c.upgradeAndServe(ctx, config)
}

func (c *Client) Healthcheck(ctx context.Context) error {
	for _, consumer := range c.cli.consumers {
		err := consumer.Healthcheck(ctx)
		if err != nil {
			return errors.WithMessage(err, "kafka consumer healthcheck")
		}
	}

	for _, publisher := range c.cli.publishers {
		err := publisher.Healthcheck(ctx)
		if err != nil {
			return errors.WithMessage(err, "kafka publisher healthcheck")
		}
	}

	return nil
}

func (c *Client) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.cli == nil {
		return
	}

	c.cli.observer.ShutdownStarted()

	closeGroup, _ := errgroup.WithContext(context.Background())
	for _, consumer := range c.cli.consumers {
		closeGroup.Go(func() error {
			err := consumer.Close()
			if err != nil {
				c.cli.observer.ClientError(err)
			}
			return nil
		})
	}

	for _, publisher := range c.cli.publishers {
		closeGroup.Go(func() error {
			err := publisher.Close()
			if err != nil {
				c.cli.observer.ClientError(err)
			}
			return nil
		})
	}

	_ = closeGroup.Wait()
	c.cli.observer.ShutdownDone()
	c.cli = nil
}

func (c *Client) upgradeAndServe(ctx context.Context, config Config) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.logger.Debug(ctx, "kafka client: received new config")

	if reflect.DeepEqual(c.prevCfg, config) {
		c.logger.Debug(ctx, "kafka client: configs are equal. skipping initialization")
		return nil
	}

	c.logger.Debug(ctx, "kafka client: initialization began")

	if c.cli != nil {
		c.Close()
		c.cli = nil
	}

	c.cli = c.upgrade(ctx, config)
	c.cli.run(ctx)

	c.cli.observer.ClientReady()

	c.prevCfg = config
	return nil
}

func (c *Client) upgrade(ctx context.Context, config Config) *KafkaClient {
	return &KafkaClient{
		publishers: config.Publishers,
		consumers:  config.Consumers,
		observer:   NewLogObserver(ctx, c.logger),
	}
}

func (c *KafkaClient) run(ctx context.Context) {
	for _, consumer := range c.consumers {
		consumer.Run(ctx)
	}
}

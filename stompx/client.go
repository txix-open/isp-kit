package stompx

import (
	"context"
	"reflect"
	"sync"

	"github.com/txix-open/isp-kit/log"
	"github.com/txix-open/isp-kit/stompx/consumer"
	"github.com/txix-open/isp-kit/stompx/publisher"
	"golang.org/x/sync/errgroup"
)

type state struct {
	consumers  []*consumer.Watcher
	publishers []*publisher.Publisher
}

type Client struct {
	locker  sync.Locker
	state   *state
	prevCfg Config
	logger  log.Logger
}

func New(logger log.Logger) *Client {
	return &Client{
		locker:  &sync.Mutex{},
		state:   nil,
		prevCfg: Config{},
		logger:  logger,
	}
}

func (c *Client) Close() error {
	c.locker.Lock()
	defer c.locker.Unlock()

	if c.state == nil {
		return nil
	}
	c.Shutdown()
	c.state = nil

	return nil
}

func (c *Client) Shutdown() {
	closeGroup, _ := errgroup.WithContext(context.Background())
	for _, consumer := range c.state.consumers {
		closeGroup.Go(func() error {
			consumer.Shutdown()
			return nil
		})
	}

	for _, publisher := range c.state.publishers {
		closeGroup.Go(func() error {
			_ = publisher.Close()
			return nil
		})
	}

	_ = closeGroup.Wait()
}

func (c *Client) Upgrade(ctx context.Context, config Config) error {
	return c.upgrade(ctx, false, config)
}

func (c *Client) UpgradeAndServe(ctx context.Context, config Config) {
	_ = c.upgrade(ctx, true, config)
}

func (c *Client) upgrade(ctx context.Context, justServe bool, newConfig Config) error {
	c.locker.Lock()
	defer c.locker.Unlock()

	c.logger.Debug(ctx, "stomp client: received new config")

	if reflect.DeepEqual(c.prevCfg, newConfig) {
		c.logger.Debug(ctx, "stomp client: configs are equal. skipping upgrade")
		return nil
	}

	c.logger.Debug(ctx, "stomp client: initialization began")

	if c.state != nil {
		c.Shutdown()
		c.state = nil
	}

	c.setNewState()

	for _, consumer := range newConfig.getConsumers() {
		if justServe {
			consumer.Serve(ctx)
		} else {
			err := consumer.Run(ctx)
			if err != nil {
				return err
			}
		}

		c.state.consumers = append(c.state.consumers, consumer)
	}
	c.state.publishers = newConfig.Publishers

	c.logger.Debug(ctx, "stomp client: initialization done")
	c.prevCfg = newConfig

	return nil
}

func (c *Client) setNewState() {
	c.state = &state{
		publishers: make([]*publisher.Publisher, 0),
		consumers:  make([]*consumer.Watcher, 0),
	}
}

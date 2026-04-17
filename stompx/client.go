// Package stompx provides a high-level wrapper over the STOMP protocol,
// implementing convenient tools for creating message consumers and publishers,
// with support for middleware, logging, retries, and management of consumer and publisher groups.
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

// Client manages a group of consumers and publishers, capable of updating
// connections and restarting when configuration changes.
type Client struct {
	locker  sync.Locker
	state   *state
	prevCfg Config
	logger  log.Logger
}

// New creates a new Client with the provided logger.
func New(logger log.Logger) *Client {
	return &Client{
		locker:  &sync.Mutex{},
		state:   nil,
		prevCfg: Config{},
		logger:  logger,
	}
}

// Close terminates all active connections.
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

// Shutdown gracefully shuts down all consumers and publishers.
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

// Upgrade updates the configuration and synchronously initializes the client
// with a guarantee that all components are ready. It returns the first error
// encountered during initialization, or nil if successful.
func (c *Client) Upgrade(ctx context.Context, config Config) error {
	return c.upgrade(ctx, false, config)
}

// UpgradeAndServe updates the configuration and restarts connections.
// It stops old connections, initializes new consumers and publishers,
// and starts message processing. This method blocks and serves indefinitely.
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

	for _, consumerCli := range newConfig.Consumers {
		if justServe {
			consumerCli.Serve(ctx)
		} else {
			err := consumerCli.Run(ctx)
			if err != nil {
				return err
			}
		}

		c.state.consumers = append(c.state.consumers, consumerCli)
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

package kafkax

import (
	"context"
	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
	"golang.org/x/sync/errgroup"
	"reflect"
	"sync"
)

// state represents the runtime state of a Client, including active publishers,
// consumers, and the lifecycle observer.
type state struct {
	publishers []*publisher.Publisher
	consumers  []consumer.Consumer
	observer   Observer
}

// run starts all configured consumers. It does not block.
func (c *state) run(ctx context.Context) {
	for _, consumer := range c.consumers {
		consumer.Run(ctx)
	}
}

// Client manages the lifecycle of Kafka publishers and consumers. It supports
// dynamic reconfiguration through UpgradeAndServe and provides health checking
// and graceful shutdown capabilities.
//
// Client is safe for concurrent use.
type Client struct {
	prevCfg Config
	state   *state
	lock    sync.Locker
	logger  log.Logger
}

// New creates a new Client instance with the provided logger.
func New(logger log.Logger) *Client {
	return &Client{
		prevCfg: Config{},
		state:   nil,
		lock:    &sync.Mutex{},
		logger:  logger,
	}
}

// UpgradeAndServe initializes or reinitializes Kafka publishers and consumers
// based on the provided configuration. If the new configuration is identical
// to the previous one, initialization is skipped. Existing consumers and
// publishers are gracefully shut down before new ones are started.
func (c *Client) UpgradeAndServe(ctx context.Context, config Config) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.logger.Debug(ctx, "kafka client: received new config")

	if reflect.DeepEqual(c.prevCfg, config) {
		c.logger.Debug(ctx, "kafka client: configs are equal. skipping initialization")
		return
	}

	c.logger.Debug(ctx, "kafka client: initialization began")

	if c.state != nil {
		c.Shutdown()
		c.state = nil
	}

	c.state = c.newState(ctx, config)
	c.state.run(ctx)

	c.state.observer.ClientReady()
	c.logger.Debug(ctx, "kafka client: initialization done")

	c.prevCfg = config
}

// Healthcheck verifies the health of all consumers and publishers. Returns an
// error if any component is unhealthy.
func (c *Client) Healthcheck(ctx context.Context) error {
	for _, consumer := range c.state.consumers {
		err := consumer.Healthcheck(ctx)
		if err != nil {
			return errors.WithMessage(err, "kafka consumer healthcheck")
		}
	}

	for _, publisher := range c.state.publishers {
		err := publisher.Healthcheck(ctx)
		if err != nil {
			return errors.WithMessage(err, "kafka publisher healthcheck")
		}
	}

	return nil
}

// Close gracefully shuts down the client and releases all resources. It is
// safe to call Close multiple times.
func (c *Client) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.state == nil {
		return
	}
	c.Shutdown()
	c.state = nil
}

// Shutdown gracefully stops all consumers and publishers, waiting for pending
// operations to complete.
func (c *Client) Shutdown() {
	c.state.observer.ShutdownStarted()

	closeGroup, _ := errgroup.WithContext(context.Background())
	for _, consumer := range c.state.consumers {
		closeGroup.Go(func() error {
			err := consumer.Close()
			if err != nil {
				c.state.observer.ClientError(err)
			}
			return nil
		})
	}

	for _, publisher := range c.state.publishers {
		closeGroup.Go(func() error {
			err := publisher.Close()
			if err != nil {
				c.state.observer.ClientError(err)
			}
			return nil
		})
	}

	_ = closeGroup.Wait()
	c.state.observer.ShutdownDone()
}

// newState creates a new state instance with the provided configuration and
// initializes the log observer.
func (c *Client) newState(ctx context.Context, config Config) *state {
	return &state{
		publishers: config.Publishers,
		consumers:  config.Consumers,
		observer:   NewLogObserver(ctx, c.logger),
	}
}

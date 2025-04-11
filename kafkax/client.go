package kafkax

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/kafkax/consumer"
	"github.com/txix-open/isp-kit/kafkax/publisher"
	"github.com/txix-open/isp-kit/log"
	"golang.org/x/sync/errgroup"
)

const sendMetricPeriod = 1 * time.Second

type state struct {
	publishers []*publisher.Publisher
	consumers  []consumer.Consumer
	observer   Observer
}

func (c *state) run(ctx context.Context) {
	for _, consumer := range c.consumers {
		consumer.Run(ctx)
	}
}

type Client struct {
	prevCfg Config
	state   *state
	lock    sync.Locker
	logger  log.Logger
}

func New(logger log.Logger) *Client {
	return &Client{
		prevCfg: Config{},
		state:   nil,
		lock:    &sync.Mutex{},
		logger:  logger,
	}
}

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

func (c *Client) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.state == nil {
		return
	}
	c.Shutdown()
	c.state = nil
}

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

func (c *Client) newState(ctx context.Context, config Config) *state {
	return &state{
		publishers: config.Publishers,
		consumers:  config.Consumers,
		observer:   NewLogObserver(ctx, c.logger),
	}
}

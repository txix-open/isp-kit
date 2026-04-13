package grmqx

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq"
	"github.com/txix-open/isp-kit/log"
)

const (
	DefaultHeartbeat   = 3 * time.Second
	DefaultDialTimeout = 5 * time.Second
)

type Client struct {
	cli     *grmq.Client
	prevCfg Config
	lock    sync.Locker
	logger  log.Logger
}

func New(logger log.Logger) *Client {
	return &Client{
		cli:     nil,
		prevCfg: Config{},
		lock:    &sync.Mutex{},
		logger:  logger,
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

// QueuesDelete removes queues on the broker by name using underlying grmq.UnsafeConnection.
func (c *Client) QueuesDelete(ctx context.Context, queueNames ...string) error {
	if len(queueNames) == 0 {
		return nil
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.cli == nil {
		return errors.New("client is not initialized")
	}
	conn := c.cli.UnsafeConnection()
	if conn == nil {
		return errors.New("rabbit mq is not connected")
	}
	ch, err := conn.Channel()
	if err != nil {
		return errors.WithMessage(err, "open channel")
	}
	defer ch.Close()

	for _, name := range queueNames {
		if name == "" {
			continue
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if _, err := ch.QueueDelete(name, false, false, false); err != nil {
			c.logger.Warn(ctx, "failed delete queue ", log.String("name", name))
		}
	}
	return nil
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

	var observer grmq.Observer

	observer = NewLogObserver(ctx, c.logger)

	if config.NewObserver != nil {
		observer = config.NewObserver(ctx, c.logger)
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
		grmq.WithObserver(observer),
	)

	if justServe {
		cli.Serve(ctx)
	} else {
		err := cli.Run(ctx)
		if err != nil {
			return err
		}
	}

	c.logger.Debug(ctx, "rmq client: initialization done")

	c.cli = cli
	c.prevCfg = config
	return nil
}

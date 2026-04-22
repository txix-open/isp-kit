// Package grmqx provides a high-level wrapper for the RabbitMQ client,
// offering automatic topology declaration,
// integration with metrics and tracing, flexible retry policies, and
// contextual logging.
package grmqx

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/txix-open/grmq"
	"github.com/txix-open/isp-kit/log"
)

const (
	// DefaultHeartbeat specifies the default heartbeat interval for RabbitMQ connections.
	DefaultHeartbeat = 3 * time.Second
	// DefaultDialTimeout specifies the default timeout for establishing connections.
	DefaultDialTimeout = 5 * time.Second
)

var (
	ErrNotExistQueue     = errors.New("queue does not exist")
	ErrQueueNameIsEmpty  = errors.New("queue name is empty")
	ErrQueuesNameIsEmpty = errors.New("queues name is empty")
)

// Client manages RabbitMQ connections and the lifecycle of consumers and publishers.
// It supports dynamic configuration updates and is safe for concurrent use.
type Client struct {
	cli     *grmq.Client
	prevCfg Config
	lock    sync.Locker
	logger  log.Logger
}

// New creates a new RabbitMQ client instance.
func New(logger log.Logger) *Client {
	return &Client{
		cli:     nil,
		prevCfg: Config{},
		lock:    &sync.Mutex{},
		logger:  logger,
	}
}

// Upgrade updates the client configuration and synchronously initializes the client,
// ensuring all components (consumers, publishers, declarations) are ready before returning.
// It blocks until the first successful session is established or an error occurs.
// Returns the first error encountered during session establishment, or nil on success.
func (c *Client) Upgrade(ctx context.Context, config Config) error {
	return c.upgrade(ctx, config, false)
}

// UpgradeAndServe updates the client configuration similarly to Upgrade, but does not wait
// for the first successful session. Errors are passed to the Observer for handling with retries.
func (c *Client) UpgradeAndServe(ctx context.Context, config Config) {
	_ = c.upgrade(ctx, config, true)
}

// Healthcheck verifies the ability to connect to the RabbitMQ broker.
// Returns an error if the client is not initialized or if the connection fails.
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

// Close terminates all connections and stops the client.
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

func (c *Client) QueueInspect(name string) (amqp091.Queue, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	ch, err := c.getChannel()
	if err != nil {
		return amqp091.Queue{}, errors.WithMessage(err, "get channel")
	}
	defer ch.Close()

	queue, err := c.queueInspect(name, ch)
	if err != nil {
		return amqp091.Queue{}, err
	}

	return queue, nil
}

// DeleteQueues deletes the specified queues from the broker.
// If an error occurs while deleting a specific queue, it is logged and processing
// continues for remaining queues. Returns immediately without error if no queue names are provided.

func (c *Client) DeleteQueues(ctx context.Context, queueNames ...string) error {
	if len(queueNames) == 0 {
		return ErrQueuesNameIsEmpty
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	ch, err := c.getChannel()
	if err != nil {
		return errors.WithMessage(err, "get channel")
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
		_, err = ch.QueueDelete(name, false, false, false)
		if err != nil {
			c.logger.Warn(ctx, "failed delete queue", log.String("name", name), log.String("error", err.Error()))
		}
	}
	return nil
}

func (c *Client) DeleteQueuesWithInspect(ctx context.Context, queueNames ...string) (map[string]error, error) {
	if len(queueNames) == 0 {
		return nil, ErrQueuesNameIsEmpty
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	rslt := make(map[string]error)
	for _, name := range queueNames {
		select {
		case <-ctx.Done():
			return rslt, ctx.Err()
		default:
		}

		ch, err := c.getChannel()
		if err != nil {
			return nil, errors.WithMessage(err, "get channel for inspect")
		}
		queue, err := c.queueInspect(name, ch)
		if err == nil {
			if queue.Consumers > 0 {
				rslt[name] = errors.Errorf("The queue in use cannot be deleted. Consumers num: %d", queue.Consumers)
				continue
			}

			if queue.Messages > 0 {
				rslt[name] = errors.Errorf("The queue containing messages cannot be deleted. Messages num: %d", queue.Messages)
				continue
			}

			rslt[name] = nil
			continue
		}
		rslt[name] = errors.WithMessage(err, "queue inspect")
		_ = ch.Close()
	}

	ch, err := c.getChannel()
	if err != nil {
		return nil, errors.WithMessage(err, "get channel for delete")
	}
	for name, errorInspect := range rslt {
		select {
		case <-ctx.Done():
			return rslt, ctx.Err()
		default:
		}

		if errorInspect == nil {
			_, deleteError := ch.QueueDelete(name, false, false, false)

			if deleteError != nil {
				c.logger.Warn(ctx, "failed delete queue", log.String("name", name), log.String("error", deleteError.Error()))
				rslt[name] = errors.WithMessage(deleteError, "queue delete")
			}
		}
	}
	_ = ch.Close()

	return rslt, nil
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

// queueInspect Using the transmitted channel "ch", it searches for a queue named "name" and returns it.
// WARNING: if there is no queue, the channel will be closed due to the features of the basic library.
// For an empty queue name, the channel will not be closed, because it will immediately return with an error.
func (c *Client) queueInspect(name string, ch *amqp091.Channel) (amqp091.Queue, error) {
	if len(name) == 0 {
		return amqp091.Queue{}, ErrQueueNameIsEmpty
	}

	queue, err := ch.QueueDeclarePassive(name, false, false, false, false, nil)
	if err != nil {
		is404err := strings.Contains(err.Error(), "Exception (404)")
		if is404err {
			return amqp091.Queue{}, ErrNotExistQueue
		}

		return amqp091.Queue{}, errors.WithMessagef(err, "inspect queue %s", name)
	}

	return queue, nil
}

// getChannel returns the active channel for the current connection.
// It is necessary to close this channel after use.
func (c *Client) getChannel() (*amqp091.Channel, error) {
	if c.cli == nil {
		return nil, errors.New("client is not initialized")
	}
	conn := c.cli.UnsafeConnection()
	if conn == nil {
		return nil, errors.New("rabbit mq is not connected")
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, errors.WithMessage(err, "open channel")
	}

	return ch, nil
}

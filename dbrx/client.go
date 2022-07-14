package dbrx

import (
	"context"
	"database/sql"
	"reflect"
	"sync"

	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/metrics"
	"github.com/integration-system/isp-kit/metrics/db_metrics"
	"github.com/pkg/errors"
)

var (
	ErrClientIsNotInitialized = errors.New("client is not initialized")
)

type Client struct {
	options []dbx.Option
	lock    *sync.RWMutex
	prevCfg dbx.Config
	cli     *dbx.Client
}

func New(opts ...dbx.Option) *Client {
	return &Client{
		options: opts,
		prevCfg: dbx.Config{},
		lock:    &sync.RWMutex{},
		cli:     nil,
	}
}
func (c *Client) Upgrade(ctx context.Context, config dbx.Config) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if reflect.DeepEqual(c.prevCfg, config) {
		return nil
	}

	cli, err := dbx.Open(ctx, config, c.options...)
	if err != nil {
		return errors.WithMessage(err, "open new client")
	}

	if c.cli != nil {
		_ = c.cli.Close()
	}

	c.cli = cli
	c.prevCfg = config

	db_metrics.Register(metrics.DefaultRegistry, c.cli.Client.DB.DB, config.Database)

	return nil
}

func (c *Client) DB() (*dbx.Client, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.db()
}

func (c *Client) Select(ctx context.Context, ptr interface{}, query string, args ...interface{}) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.Select(ctx, ptr, query, args...)
}

func (c *Client) SelectRow(ctx context.Context, ptr interface{}, query string, args ...interface{}) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.SelectRow(ctx, ptr, query, args...)
}

func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	cli, err := c.db()
	if err != nil {
		return nil, err
	}
	return cli.Exec(ctx, query, args...)
}

func (c *Client) ExecNamed(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	cli, err := c.db()
	if err != nil {
		return nil, err
	}
	return cli.ExecNamed(ctx, query, arg)
}

func (c *Client) RunInTransaction(ctx context.Context, txFunc db.TxFunc, opts ...db.TxOption) error {
	c.lock.RLock()
	defer c.lock.RUnlock()

	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.RunInTransaction(ctx, txFunc, opts...)
}

func (c *Client) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.prevCfg = dbx.Config{}
	cli := c.cli
	c.cli = nil
	if cli != nil {
		return cli.Close()
	}
	return nil
}

func (c *Client) db() (*dbx.Client, error) {
	if c.cli == nil {
		return nil, ErrClientIsNotInitialized
	}
	return c.cli, nil
}

package dbrx

import (
	"context"
	"database/sql"
	"github.com/txix-open/isp-kit/log"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/db_metrics"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/sql_tracing"
)

var (
	ErrClientIsNotInitialized = errors.New("client is not initialized")
)

type Client struct {
	options []dbx.Option
	prevCfg *atomic.Value
	cli     *atomic.Pointer[dbx.Client]
	logger  log.Logger
}

func New(logger log.Logger, opts ...dbx.Option) *Client {
	prevCfg := &atomic.Value{}
	prevCfg.Store(dbx.Config{})
	return &Client{
		options: opts,
		prevCfg: prevCfg,
		cli:     &atomic.Pointer[dbx.Client]{},
		logger:  logger,
	}
}
func (c *Client) Upgrade(ctx context.Context, config dbx.Config) error {
	c.logger.Debug(ctx, "db client: received new config")

	if reflect.DeepEqual(c.prevCfg.Load(), config) {
		c.logger.Debug(ctx, "db client: configs are equal. skipping initialization")
		return nil
	}

	c.logger.Debug(ctx, "db client: initialization began")

	metricsTracer := sql_metrics.NewTracer(metrics.DefaultRegistry)
	tracingConfig := sql_tracing.NewConfig()
	tracingConfig.EnableStatement = true
	opts := append([]dbx.Option{
		dbx.WithQueryTracer(metricsTracer, tracingConfig.QueryTracer()),
	}, c.options...)

	newCli, err := dbx.Open(ctx, config, opts...)
	if err != nil {
		return errors.WithMessage(err, "open new client")
	}

	oldCli := c.cli.Swap(newCli)
	if oldCli != nil {
		_ = oldCli.Close()
	}

	c.logger.Debug(ctx, "db client: initialization done")

	c.prevCfg.Store(config)

	db_metrics.Register(metrics.DefaultRegistry, newCli.Client.DB.DB, config.Database)

	return nil
}

func (c *Client) DB() (*dbx.Client, error) {
	return c.db()
}

func (c *Client) Select(ctx context.Context, ptr any, query string, args ...any) error {
	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.Select(ctx, ptr, query, args...)
}

func (c *Client) SelectRow(ctx context.Context, ptr any, query string, args ...any) error {
	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.SelectRow(ctx, ptr, query, args...)
}

func (c *Client) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	cli, err := c.db()
	if err != nil {
		return nil, err
	}
	return cli.Exec(ctx, query, args...)
}

func (c *Client) ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error) {
	cli, err := c.db()
	if err != nil {
		return nil, err
	}
	return cli.ExecNamed(ctx, query, arg)
}

func (c *Client) RunInTransaction(ctx context.Context, txFunc db.TxFunc, opts ...db.TxOption) error {
	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.RunInTransaction(ctx, txFunc, opts...)
}

func (c *Client) Close() error {
	c.logger.Debug(context.Background(), "db client: call close")
	c.prevCfg.Store(dbx.Config{})
	oldCli := c.cli.Swap(nil)
	if oldCli != nil {
		return oldCli.Close()
	}
	return nil
}

func (c *Client) Healthcheck(ctx context.Context) error {
	cli, err := c.db()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	_, err = cli.Exec(ctx, "SELECT 1")
	if err != nil {
		return errors.WithMessage(err, "exec")
	}
	return nil
}

func (c *Client) db() (*dbx.Client, error) {
	oldCli := c.cli.Load()
	if oldCli == nil {
		return nil, ErrClientIsNotInitialized
	}
	return oldCli, nil
}

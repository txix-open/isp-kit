// Package dbrx provides a dynamic database client that supports runtime
// configuration updates. It wraps dbx.Client with atomic pointer access,
// automatic metrics and tracing integration, and hot-reload capability.
package dbrx

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/txix-open/isp-kit/log"

	"github.com/pkg/errors"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/metrics"
	"github.com/txix-open/isp-kit/metrics/db_metrics"
	"github.com/txix-open/isp-kit/metrics/sql_metrics"
	"github.com/txix-open/isp-kit/observability/tracing/sql_tracing"
)

// ErrClientIsNotInitialized is returned when the client has not been initialized.
var (
	ErrClientIsNotInitialized = errors.New("client is not initialized")
)

const (
	healthcheckTimeout = 500 * time.Millisecond
)

// Client provides dynamic database client management with hot-reload support.
// It wraps a dbx.Client and allows configuration updates at runtime.
// It is safe for concurrent use.
type Client struct {
	options []dbx.Option
	prevCfg *atomic.Value
	cli     *atomic.Pointer[dbx.Client]
	logger  log.Logger
}

// New creates a new Client with the provided logger and options.
// The client is not initialized until Upgrade is called.
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

// Upgrade initializes or reinitializes the database client with the provided configuration.
// If the new configuration is identical to the previous one, initialization is skipped.
// It automatically adds metrics tracing, schema creation, and application name options.
// Returns an error if the connection fails or if the client is in read-only mode.
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
		dbx.WithCreateSchema(true),
		dbx.WithApplicationName(os.Args[0]),
	}, c.options...)

	newCli, err := dbx.Open(ctx, config, opts...)
	if err != nil {
		return errors.WithMessage(err, "open new client")
	}

	readOnly, err := newCli.IsReadOnly(ctx)
	if err != nil {
		return errors.WithMessage(err, "check is new cli read only")
	}
	if readOnly {
		c.logger.Info(ctx, "db client: connection is in read-only mode")
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

// DB returns the underlying database client.
// Returns an error if the client has not been initialized.
func (c *Client) DB() (*dbx.Client, error) {
	return c.db()
}

// Select executes a query that returns multiple rows and scans them into the provided pointer.
// Returns an error if the client is not initialized or if the query fails.
func (c *Client) Select(ctx context.Context, ptr any, query string, args ...any) error {
	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.Select(ctx, ptr, query, args...)
}

// SelectRow executes a query that returns a single row and scans it into the provided pointer.
// Returns an error if the client is not initialized or if the query fails.
func (c *Client) SelectRow(ctx context.Context, ptr any, query string, args ...any) error {
	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.SelectRow(ctx, ptr, query, args...)
}

// Exec executes a query that does not return rows.
// Returns an error if the client is not initialized or if the query fails.
func (c *Client) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	cli, err := c.db()
	if err != nil {
		return nil, err
	}
	return cli.Exec(ctx, query, args...)
}

// ExecNamed executes a named-parameter query that does not return rows.
// Returns an error if the client is not initialized or if the query fails.
func (c *Client) ExecNamed(ctx context.Context, query string, arg any) (sql.Result, error) {
	cli, err := c.db()
	if err != nil {
		return nil, err
	}
	return cli.ExecNamed(ctx, query, arg)
}

// RunInTransaction executes the provided function within a database transaction.
// Returns an error if the client is not initialized or if the transaction fails.
func (c *Client) RunInTransaction(ctx context.Context, txFunc db.TxFunc, opts ...db.TxOption) error {
	cli, err := c.db()
	if err != nil {
		return err
	}
	return cli.RunInTransaction(ctx, txFunc, opts...)
}

// Close closes the database connection and resets the client configuration.
// Returns an error if the underlying client fails to close.
func (c *Client) Close() error {
	c.logger.Debug(context.Background(), "db client: call close")
	c.prevCfg.Store(dbx.Config{})
	oldCli := c.cli.Swap(nil)
	if oldCli != nil {
		return oldCli.Close()
	}
	return nil
}

// Healthcheck verifies that the database connection is alive.
// Executes a simple query with a 500ms timeout.
// Returns an error if the client is not initialized or if the query fails.
func (c *Client) Healthcheck(ctx context.Context) error {
	cli, err := c.db()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, healthcheckTimeout)
	defer cancel()
	_, err = cli.Exec(ctx, "SELECT 1")
	if err != nil {
		return errors.WithMessage(err, "exec")
	}
	return nil
}

// db returns the current database client or an error if not initialized.
func (c *Client) db() (*dbx.Client, error) {
	oldCli := c.cli.Load()
	if oldCli == nil {
		return nil, ErrClientIsNotInitialized
	}
	return oldCli, nil
}

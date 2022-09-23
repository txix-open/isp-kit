package dbx

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/dbx/migration"
	"github.com/pkg/errors"
)

var (
	defaultMaxOpenConn = runtime.NumCPU() * 10
)

type Client struct {
	*db.Client

	withMigration bool
	migrationDir  string
}

func Open(ctx context.Context, config Config, opts ...Option) (*Client, error) {
	cli := &Client{}
	for _, opt := range opts {
		opt(cli)
	}

	if config.Schema != "public" && config.Schema != "" {
		err := createSchema(ctx, config)
		if err != nil {
			return nil, errors.WithMessage(err, "create schema")
		}
	}

	dbCli, err := db.Open(ctx, config.Dsn())
	if err != nil {
		return nil, errors.WithMessage(err, "open db")
	}

	maxOpenConn := defaultMaxOpenConn
	if config.MaxOpenConn > 0 {
		maxOpenConn = config.MaxOpenConn
	}
	maxIdleConns := maxOpenConn / 2
	if maxIdleConns < 2 {
		maxIdleConns = 2
	}
	dbCli.SetMaxOpenConns(maxOpenConn)
	dbCli.SetMaxIdleConns(maxIdleConns)
	dbCli.SetConnMaxIdleTime(90 * time.Second)

	if cli.withMigration {
		err = migration.NewRunner(dbCli.DB.DB, cli.migrationDir).Run()
		if err != nil {
			return nil, errors.WithMessage(err, "migration run")
		}
	}

	cli.Client = dbCli
	return cli, nil
}

func createSchema(ctx context.Context, config Config) error {
	schema := config.Schema

	config.Schema = ""
	dbCli, err := db.Open(ctx, config.Dsn())
	if err != nil {
		return errors.WithMessage(err, "open db")
	}

	_, err = dbCli.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
	if err != nil {
		return errors.WithMessage(err, "exec query")
	}

	err = dbCli.Close()
	if err != nil {
		return errors.WithMessage(err, "close db")
	}
	return nil
}

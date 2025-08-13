package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/txix-open/isp-kit/db"
)

// nolint:gochecknoglobals
var (
	defaultMaxOpenConn = runtime.NumCPU() * 10
)

const (
	minIdleConns       = 2
	connMaxIdleTimeout = 90 * time.Second
)

type MigrationRunner interface {
	Run(ctx context.Context, db *sql.DB, gooseOpts ...goose.ProviderOption) error
}

type Client struct {
	*db.Client

	migrationRunner MigrationRunner
	queryTraces     []pgx.QueryTracer
	createSchema    bool
	applicationName string
}

// nolint:cyclop,nonamedreturns
func Open(ctx context.Context, config Config, opts ...Option) (cli *Client, err error) {
	cli = &Client{}
	for _, opt := range opts {
		opt(cli)
	}

	dbCli, err := db.Open(ctx, config.Dsn(cli.applicationName), db.WithQueryTracer(cli.queryTraces...))
	if err != nil {
		return nil, errors.WithMessage(err, "open db")
	}
	defer func() {
		if err != nil {
			_ = dbCli.Close()
		}
	}()

	maxOpenConn := defaultMaxOpenConn
	if config.MaxOpenConn > 0 {
		maxOpenConn = config.MaxOpenConn
	}
	maxIdleConns := max(maxOpenConn/minIdleConns, minIdleConns)
	dbCli.SetMaxOpenConns(maxOpenConn)
	dbCli.SetMaxIdleConns(maxIdleConns)
	dbCli.SetConnMaxIdleTime(connMaxIdleTimeout)

	isReadOnly, err := dbCli.IsReadOnly(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "check is read only connection")
	}

	isCustomSchema := config.Schema != "public" && config.Schema != ""
	if !isReadOnly && cli.createSchema && isCustomSchema {
		_, err = dbCli.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", config.Schema))
		if err != nil {
			return nil, errors.WithMessage(err, "exec create schema query")
		}
	}

	if isCustomSchema {
		err = checkSchemaExistence(ctx, config.Schema, dbCli)
		if err != nil {
			return nil, errors.WithMessage(err, "check schema existence")
		}
	}

	if !isReadOnly && cli.migrationRunner != nil {
		err = cli.migrationRunner.Run(ctx, dbCli.DB.DB)
		if err != nil {
			return nil, errors.WithMessage(err, "migration run")
		}
	}

	cli.Client = dbCli
	return cli, nil
}

func checkSchemaExistence(ctx context.Context, schema string, dbCli *db.Client) error {
	query := `SELECT EXISTS (
		SELECT 1 FROM pg_namespace WHERE nspname = $1
	)`
	var exists bool
	err := dbCli.QueryRowContext(ctx, query, schema).Scan(&exists)
	if err != nil {
		return errors.WithMessage(err, "check schema existence")
	}
	if !exists {
		return errors.Errorf("schema '%s' does not exist", schema)
	}
	return nil
}

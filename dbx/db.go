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
}

func Open(ctx context.Context, config Config, opts ...Option) (*Client, error) {
	cli := &Client{}
	for _, opt := range opts {
		opt(cli)
	}

	isCustomSchema := config.Schema != "public" && config.Schema != ""
	if cli.createSchema && isCustomSchema {
		err := createSchema(ctx, config)
		if err != nil {
			return nil, errors.WithMessage(err, "create schema")
		}
	}

	dbCli, err := db.Open(ctx, config.Dsn(), db.WithQueryTracer(cli.queryTraces...))
	if err != nil {
		return nil, errors.WithMessage(err, "open db")
	}

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

	if isCustomSchema {
		err := checkSchemaExistence(ctx, config.Schema, dbCli)
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

func createSchema(ctx context.Context, config Config) error {
	schema := config.Schema

	config.Schema = ""
	dbCli, err := db.Open(ctx, config.Dsn())
	if err != nil {
		return errors.WithMessage(err, "open db")
	}
	isReadOnly, err := dbCli.IsReadOnly(ctx)
	if err != nil {
		return errors.WithMessage(err, "check is read only connection")
	}
	if !isReadOnly {
		_, err = dbCli.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schema))
		if err != nil {
			return errors.WithMessage(err, "exec query")
		}
	}

	err = dbCli.Close()
	if err != nil {
		return errors.WithMessage(err, "close db")
	}
	return nil
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

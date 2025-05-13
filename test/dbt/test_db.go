package dbt

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/db"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/test"
)

type TestDb struct {
	*dbx.Client
	must   must
	schema string
}

// nolint:mnd
func New(t *test.Test, opts ...dbx.Option) *TestDb {
	dbConfig := Config(t)
	ctx, cancel := context.WithTimeout(t.T().Context(), 5*time.Second)
	defer cancel()
	cli, err := dbx.Open(ctx, dbConfig, opts...)
	t.Assert().NoError(err, "open test db cli, schema: %s", dbConfig.Schema)

	db := &TestDb{
		Client: cli,
		schema: dbConfig.Schema,
		must: must{
			db:     cli.Client,
			assert: t.Assert(),
		},
	}
	t.T().Cleanup(func() {
		err := db.Close()
		t.Assert().NoError(err)
	})
	return db
}

func (db *TestDb) DB() (*dbx.Client, error) {
	return db.Client, nil
}

func (db *TestDb) Must() must { // for test purposes
	return db.must
}

func (db *TestDb) Schema() string {
	return db.schema
}

func (db *TestDb) Close() error {
	_, err := db.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA %s CASCADE", db.schema))
	if err != nil {
		return errors.WithMessage(err, "drop schema")
	}
	err = db.Client.Close()
	return errors.WithMessage(err, "close db")
}

type must struct {
	db     *db.Client
	assert *require.Assertions
}

func (m must) Exec(query string, args ...any) sql.Result {
	res, err := m.db.Exec(context.Background(), query, args...)
	m.assert.NoError(err)
	return res
}

func (m must) Select(resultPtr any, query string, args ...any) {
	err := m.db.Select(context.Background(), resultPtr, query, args...)
	m.assert.NoError(err)
}

func (m must) SelectRow(resultPtr any, query string, args ...any) {
	err := m.db.SelectRow(context.Background(), resultPtr, query, args...)
	m.assert.NoError(err)
}

func (m must) ExecNamed(query string, arg any) sql.Result {
	res, err := m.db.ExecNamed(context.Background(), query, arg)
	m.assert.NoError(err)
	return res
}

func (m must) Count(query string, args ...any) int {
	value := 0
	m.SelectRow(&value, query, args...)
	return value
}

func Config(t *test.Test) dbx.Config {
	cfg := t.Config()
	schema := fmt.Sprintf("test_%s", t.Id()) // name must start from none digit
	dbConfig := dbx.Config{
		Host:        cfg.Optional().String("PG_HOST", "127.0.0.1"),
		Port:        cfg.Optional().Int("PG_PORT", 5432), // nolint:mnd
		Database:    cfg.Optional().String("PG_DB", "test"),
		Username:    cfg.Optional().String("PG_USER", "test"),
		Password:    cfg.Optional().String("PG_PASS", "test"),
		Schema:      schema,
		MaxOpenConn: runtime.NumCPU(),
	}
	return dbConfig
}

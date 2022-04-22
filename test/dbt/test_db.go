package dbt

import (
	"context"
	"database/sql"
	"fmt"
	"runtime"
	"time"

	"github.com/integration-system/isp-kit/db"
	"github.com/integration-system/isp-kit/dbx"
	"github.com/integration-system/isp-kit/test"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

type TestDb struct {
	*dbx.Client
	must   must
	schema string
}

func New(t *test.Test, opts ...dbx.Option) *TestDb {
	cfg := t.Config()
	schema := fmt.Sprintf("test_%s", t.Id()) //name must start from none digit
	dbConfig := dbx.Config{
		Host:        cfg.Optional().String("PG_HOST", "127.0.0.1"),
		Port:        cfg.Optional().Int("PG_PORT", 5432),
		Database:    cfg.Optional().String("PG_DB", "test"),
		Username:    cfg.Optional().String("PG_USER", "test"),
		Password:    cfg.Optional().String("PG_PASS", "test"),
		Schema:      schema,
		MaxOpenConn: runtime.NumCPU(),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cli, err := dbx.Open(ctx, dbConfig, opts...)
	t.Assert().NoError(err, errors.WithMessagef(err, "open test db cli, schema: %s", schema))

	db := &TestDb{
		Client: cli,
		schema: schema,
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

func (db *TestDb) Must() *must { //for test purposes
	return &db.must
}

func (db *TestDb) Close() error {
	_, err := db.Exec(context.Background(), fmt.Sprintf("DROP SCHEMA %s CASCADE", db.schema))
	if err != nil {
		return errors.WithMessage(err, "drop schema")
	}
	err = db.DB.Close()
	return errors.WithMessage(err, "close db")
}

type must struct {
	db     *db.Client
	assert *require.Assertions
}

func (m must) Exec(query string, args ...interface{}) sql.Result {
	res, err := m.db.Exec(context.Background(), query, args...)
	m.assert.NoError(err)
	return res
}

func (m must) Select(resultPtr interface{}, query string, args ...interface{}) {
	err := m.db.Select(context.Background(), resultPtr, query, args...)
	m.assert.NoError(err)
}

func (m must) SelectRow(resultPtr interface{}, query string, args ...interface{}) {
	err := m.db.SelectRow(context.Background(), resultPtr, query, args...)
	m.assert.NoError(err)
}

func (m must) ExecNamed(query string, arg interface{}) sql.Result {
	res, err := m.db.ExecNamed(context.Background(), query, arg)
	m.assert.NoError(err)
	return res
}

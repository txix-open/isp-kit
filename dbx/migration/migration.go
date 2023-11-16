package migration

import (
	"database/sql"
	"os"

	"github.com/integration-system/isp-kit/log"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
)

type Runner struct {
	migrationDir string
}

func NewRunner(migrationDir string, logger log.Logger) Runner {
	if logger != nil {
		goose.SetLogger(newLogger(logger))
	}
	return Runner{
		migrationDir: migrationDir,
	}
}

func (r Runner) Run(db *sql.DB) error {
	_, err := os.Stat(r.migrationDir)
	if err != nil {
		return errors.WithMessage(err, "get file info")
	}

	err = goose.Version(db, r.migrationDir)
	if err != nil {
		return errors.WithMessage(err, "print goose version")
	}

	err = goose.Status(db, r.migrationDir)
	if err != nil {
		return errors.WithMessage(err, "print goose status")
	}

	err = goose.Up(db, r.migrationDir)
	if err != nil {
		return errors.WithMessage(err, "run goose up")
	}

	return nil
}

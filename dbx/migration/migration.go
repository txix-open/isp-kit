package migration

import (
	"database/sql"
	"os"

	"github.com/pkg/errors"
	"github.com/pressly/goose"
)

type Runner struct {
	db           *sql.DB
	migrationDir string
}

func NewRunner(db *sql.DB, migrationDir string) Runner {
	return Runner{
		db:           db,
		migrationDir: migrationDir,
	}
}

func (r Runner) Run() error {
	_, err := os.Stat(r.migrationDir)
	if err != nil {
		return errors.WithMessage(err, "get file info")
	}

	err = goose.Version(r.db, r.migrationDir)
	if err != nil {
		return errors.WithMessage(err, "print goose version")
	}

	err = goose.Status(r.db, r.migrationDir)
	if err != nil {
		return errors.WithMessage(err, "print goose status")
	}

	err = goose.Run("up", r.db, r.migrationDir)
	if err != nil {
		return errors.WithMessage(err, "complete goose command")
	}

	return nil
}

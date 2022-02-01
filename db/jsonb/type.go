package jsonb

import (
	"database/sql/driver"

	"github.com/jackc/pgtype"
)

type Type []byte

func (t *Type) Scan(src interface{}) error {
	jsonb := pgtype.JSONB{}
	err := jsonb.Scan(src)
	if err != nil {
		return err
	}
	if jsonb.Status == pgtype.Null {
		return nil
	}
	*t = jsonb.Bytes
	return nil
}

func (t Type) Value() (driver.Value, error) {
	if t == nil {
		return pgtype.JSONB{Status: pgtype.Null}.Value()
	}
	return pgtype.JSONB{Status: pgtype.Present, Bytes: t}.Value()
}

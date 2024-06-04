package jsonb_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"gitlab.txix.ru/isp/isp-kit/db/jsonb"
	"gitlab.txix.ru/isp/isp-kit/test"
	"gitlab.txix.ru/isp/isp-kit/test/dbt"
)

type record struct {
	Id   int64
	Data jsonb.Type
}

type someData struct {
	Value int
}

func TestType(t *testing.T) {
	test, assert := test.New(t)
	db := dbt.New(test)
	createTable := `
create table data (
	id serial8 primary key,
    data jsonb
)
`
	db.Must().Exec(createTable)
	data, err := json.Marshal(someData{123})
	assert.NoError(err)
	expected := record{Id: 1, Data: data}
	db.Must().Exec("insert into data values ($1, $2)", expected.Id, expected.Data)
	actual := record{}
	db.Must().SelectRow(&actual, "select * from data where id = $1", 1)
	actual.Data = bytes.ReplaceAll(actual.Data, []byte(" "), []byte("")) // pg add extra spaces
	assert.EqualValues(expected, actual)

	expected = record{Id: 2, Data: nil}
	db.Must().Exec("insert into data values ($1, $2)", expected.Id, expected.Data)
	actual = record{}
	db.Must().SelectRow(&actual, "select * from data where id = $1", 2)
	assert.EqualValues(expected, actual)
}

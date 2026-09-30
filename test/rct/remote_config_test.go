package rct_test

import (
	"testing"

	"github.com/txix-open/isp-kit/test/rct"
)

type Child struct {
	Value string `validate:"required"`
}

type Config struct {
	String  string `validate:"required"`
	Integer int
	Child   Child
}

func Test(t *testing.T) {
	t.Parallel()
	rct.Test(t, "config.json", Config{})
}

package rct_test

import (
	"testing"

	"github.com/integration-system/isp-kit/test/rct"
)

type Child struct {
	Value string `valid:"required"`
}

type Config struct {
	String  string `valid:"required"`
	Integer int
	Child   Child
}

func Test(t *testing.T) {
	rct.Test(t, "config.json", Config{})
}

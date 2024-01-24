package rct_test

import (
	"testing"

	"github.com/integration-system/isp-kit/test/rct"
	"github.com/stretchr/testify/assert"
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
	rct.Test(t, "config.json", Config{})
}

func TestFindTag(t *testing.T) {
	tag := "validate"
	assert.True(t, rct.FindTag(Config{}, tag))
	assert.True(t, rct.FindTag(&Config{}, tag))
	assert.True(t, rct.FindTag[*Config](nil, tag))
	type s struct {
		Cfg Config
	}
	assert.True(t, rct.FindTag(s{}, tag))
	assert.True(t, rct.FindTag(map[string]Config{}, tag))
	assert.True(t, rct.FindTag(map[string]*Config{}, tag))
	assert.True(t, rct.FindTag([]*Config{}, tag))
	assert.True(t, rct.FindTag([]Config{}, tag))
	assert.True(t, rct.FindTag[[]map[string][]*s](nil, tag))
}

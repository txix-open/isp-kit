package rct_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestFindTag(t *testing.T) {
	t.Parallel()

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

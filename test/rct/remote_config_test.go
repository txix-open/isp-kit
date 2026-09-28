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

type recursiveNode struct {
	Name  string `validate:"required"`
	Items *recursiveNode
	Props map[string]*recursiveNode
	Vars  []map[string]*recursiveNode
}

type nonRecursiveNode struct {
	Items *nonRecursiveNode `validate:"required"`
}

func TestFindTagRecursive(t *testing.T) {
	t.Parallel()

	tag := "validate"
	assert.True(t, rct.FindTag(recursiveNode{}, tag))
	assert.True(t, rct.FindTag(&recursiveNode{}, tag))
	assert.True(t, rct.FindTag[*recursiveNode](nil, tag))
	assert.True(t, rct.FindTag(map[string]*recursiveNode{}, tag))
	assert.False(t, rct.FindTag(recursiveNode{}, "nonexistent"))
	assert.False(t, rct.FindTag(nonRecursiveNode{}, "nonexistent"))
}

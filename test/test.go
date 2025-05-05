package test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/config"
	"github.com/txix-open/isp-kit/log"
)

const testIdLength = 4

type Test struct {
	id         string
	cfg        *config.Config
	logger     log.Logger
	t          *testing.T
	assertions *require.Assertions
}

func New(t *testing.T) (*Test, *require.Assertions) {
	t.Helper()

	assert := require.New(t)
	cfg, err := config.New()
	assert.NoError(err)

	logger, err := log.New(log.WithDevelopmentMode(), log.WithLevel(log.DebugLevel))
	assert.NoError(err)

	idBytes := make([]byte, testIdLength)
	_, err = rand.Read(idBytes)
	assert.NoError(err)
	return &Test{
		id:         hex.EncodeToString(idBytes),
		t:          t,
		cfg:        cfg,
		logger:     logger,
		assertions: assert,
	}, assert
}

func (t *Test) Config() *config.Config {
	return t.cfg
}

// nolint:ireturn
func (t *Test) Logger() log.Logger {
	return t.logger
}

func (t *Test) Assert() *require.Assertions {
	return t.assertions
}

func (t *Test) Id() string {
	return t.id
}

func (t *Test) T() *testing.T {
	return t.t
}

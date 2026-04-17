// Package test provides a test helper that wraps testing.T and automatically
// initializes configuration and logger for unit and integration tests.
//
// The Test struct serves as a central container for commonly used test
// dependencies including configuration, logger, testing.T context, and
// require.Assertions for assertions.
package test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/txix-open/isp-kit/config"
	"github.com/txix-open/isp-kit/log"
)

const (
	testIdLength = 4
)

// Test is a test helper that wraps testing.T and provides access to
// initialized configuration, logger, and assertion utilities.
type Test struct {
	id         string
	cfg        *config.Config
	logger     log.Logger
	t          *testing.T
	assertions *require.Assertions
}

// New creates and returns a Test instance and require.Assertions.
//
// It automatically initializes configuration and logger, generates a
// unique test identifier, and creates require assertions. The function
// marks the calling test as a helper.
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

// Config returns the configuration instance associated with the test.
func (t *Test) Config() *config.Config {
	return t.cfg
}

// Logger returns the logger instance associated with the test.
// The logger is configured in development mode with debug level.
//
// nolint:ireturn
func (t *Test) Logger() log.Logger {
	return t.logger
}

// Assert returns the require.Assertions instance for performing
// assertions within the test.
func (t *Test) Assert() *require.Assertions {
	return t.assertions
}

// Id returns the unique identifier for the test as a hexadecimal string.
func (t *Test) Id() string {
	return t.id
}

// T returns the underlying testing.T instance.
func (t *Test) T() *testing.T {
	return t.t
}

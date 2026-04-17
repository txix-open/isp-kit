// Package rc provides a configuration management system with support for override merging,
// validation, and hot-reloading capabilities. It uses JSON as the internal format and
// supports hierarchical configuration through path delimiters.
package rc

import (
	"maps"
	"sync"

	"github.com/pkg/errors"
	"github.com/txix-open/bellows"
	"github.com/txix-open/isp-kit/json"
)

// Validator defines an interface for validating configuration values.
// Implementations should return an error if the provided value is invalid.
type Validator interface {
	ValidateToError(value any) error
}

// Config manages configuration state with support for override merging and validation.
// It is safe for concurrent use.
type Config struct {
	prevConfig     []byte
	overrideConfig []byte
	delim          string
	validator      Validator
	lock           sync.Locker
}

// New creates a new Config instance with the provided validator and override data.
// The overrideData is merged with new configuration data during upgrades.
// The default delimiter for hierarchical paths is "~".
func New(validator Validator, overrideData []byte) *Config {
	return &Config{
		prevConfig:     nil,
		overrideConfig: overrideData,
		delim:          "~",
		validator:      validator,
		lock:           &sync.Mutex{},
	}
}

// Upgrade processes new configuration data by merging it with overrides,
// unmarshaling into the provided config pointer, and validating the result.
// If validation succeeds, it stores the new config as the previous config.
// The prevConfigPtr is populated with the previous configuration state if available.
// Returns an error if merging, unmarshaling, or validation fails.
func (c *Config) Upgrade(data []byte, newConfigPtr any, prevConfigPtr any) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	newConfig, err := c.mergeWithOverride(data)
	if err != nil {
		return errors.WithMessage(err, "merge with override new config")
	}

	err = json.Unmarshal(newConfig, newConfigPtr)
	if err != nil {
		return errors.WithMessage(err, "unmarshal new config")
	}

	err = c.validator.ValidateToError(newConfigPtr)
	if err != nil {
		return errors.WithMessage(err, "validate config")
	}

	if len(c.prevConfig) > 0 {
		err = json.Unmarshal(c.prevConfig, prevConfigPtr)
		if err != nil {
			return errors.WithMessage(err, "unmarshal previous config")
		}
	}

	c.prevConfig = newConfig

	return nil
}

// mergeWithOverride merges the provided config data with the override configuration.
// It flattens both configs, applies overrides, and expands the result back to hierarchical form.
// Returns the merged configuration as JSON bytes, or an error if merging fails.
func (c *Config) mergeWithOverride(data []byte) ([]byte, error) {
	config := make(map[string]any)
	err := json.Unmarshal(data, &config)
	if err != nil {
		return nil, errors.WithMessage(err, "unmarshal config")
	}

	overrideData := make(map[string]any)
	if len(c.overrideConfig) > 0 {
		err = json.Unmarshal(c.overrideConfig, &overrideData)
		if err != nil {
			return nil, errors.WithMessage(err, "unmarshal override data")
		}
	}

	config = bellows.Flatten(config, bellows.WithSep(c.delim))
	overrideData = bellows.Flatten(overrideData, bellows.WithSep(c.delim))
	maps.Copy(config, overrideData)

	result := bellows.Expand(config, bellows.WithSep(c.delim))
	if result == nil {
		result = make(map[string]any)
	}
	config, ok := result.(map[string]any)
	if !ok {
		return nil, errors.WithMessagef(err, "unexpected type from bellows, expected map, got %T", result)
	}

	data, err = json.Marshal(config)
	if err != nil {
		return nil, errors.WithMessage(err, "marshal config")
	}

	return data, nil
}

// Upgrade is a generic wrapper for Config.Upgrade that returns typed configuration structs.
// It unmarshals the new and previous configurations into the generic types T.
// Returns the new configuration, the previous configuration, and any error encountered.
func Upgrade[T any](rc *Config, data []byte) (newCfg T, prevCfg T, err error) {
	err = rc.Upgrade(data, &newCfg, &prevCfg)
	return newCfg, prevCfg, err
}

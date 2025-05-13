package rc

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/txix-open/bellows"
	"github.com/txix-open/isp-kit/json"
)

type Validator interface {
	ValidateToError(value any) error
}

type Config struct {
	prevConfig     []byte
	overrideConfig []byte
	delim          string
	validator      Validator
	lock           sync.Locker
}

func New(validator Validator, overrideData []byte) *Config {
	return &Config{
		prevConfig:     nil,
		overrideConfig: overrideData,
		delim:          "~",
		validator:      validator,
		lock:           &sync.Mutex{},
	}
}

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
	for k, v := range overrideData {
		config[k] = v
	}
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

// nolint:ireturn,nonamedreturns
func Upgrade[T any](rc *Config, data []byte) (newCfg T, prevCfg T, err error) {
	err = rc.Upgrade(data, &newCfg, &prevCfg)
	return newCfg, prevCfg, err
}

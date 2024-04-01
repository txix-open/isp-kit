package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/txix-open/bellows"
)

type Validator interface {
	ValidateToError(value any) error
}

type Config struct {
	config    map[string]string
	optional  Optional
	mandatory Mandatory

	envPrefix    string
	validator    Validator
	extraSources []Source
}

func New(opts ...Option) (*Config, error) {
	config := map[string]string{}
	mandatory := Mandatory{config: config}
	optional := Optional{m: mandatory}
	cfg := &Config{
		config:    config,
		mandatory: mandatory,
		optional:  optional,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	for _, source := range cfg.extraSources {
		extraConfig, err := source.Config()
		if err != nil {
			return nil, errors.WithMessagef(err, "read extra source, %T", source)
		}
		for key, value := range extraConfig {
			config[normalizeKey(key)] = value
		}
	}

	for _, pairs := range os.Environ() {
		parts := strings.Split(pairs, "=")
		key := normalizeKey(parts[0])
		prefix := normalizeKey(cfg.envPrefix)
		if prefix != "" && !strings.HasPrefix(key, prefix) {
			continue
		}
		key = key[len(prefix):]
		config[key] = strings.Join(parts[1:], "")
	}

	return cfg, nil
}

func (c *Config) Set(key string, value any) {
	c.config[key] = fmt.Sprintf("%v", value)
}

func (c *Config) Delete(key string) {
	delete(c.config, key)
}

func (c *Config) Mandatory() Mandatory {
	return c.mandatory
}

func (c *Config) Optional() Optional {
	return c.optional
}

func (c *Config) Read(ptr any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(mapstructure.StringToTimeDurationHookFunc()),
		WeaklyTypedInput: true,
		Result:           ptr,
		Squash:           true,
	})
	if err != nil {
		return errors.WithMessage(err, "mapstructure new decoder")
	}

	expanded := make(map[string]any)
	for key, value := range c.config {
		expanded[key] = value
	}
	toDecode := bellows.Expand(expanded)
	err = decoder.Decode(toDecode)
	if err != nil {
		return errors.WithMessage(err, "decode config")
	}

	if c.validator != nil {
		err := c.validator.ValidateToError(ptr)
		if err != nil {
			return errors.WithMessage(err, "validate config")
		}
	}

	return nil
}

func normalizeKey(key string) string {
	return strings.ToLower(key)
}

func get[T any](config map[string]string, key string, valueMapper func(value string) (T, error)) (T, error) {
	var ret T
	value, ok := config[normalizeKey(key)]
	if !ok {
		return ret, errors.Errorf("%s is expected in config", key)
	}
	mapped, err := valueMapper(value)
	if err != nil {
		return ret, err
	}

	return mapped, nil
}

// Package config provides a flexible configuration management system with support for
// multiple sources, environment variable overrides, and type-safe value retrieval.
//
// The package supports reading configuration from YAML files, environment variables,
// and custom sources. Configuration values can be accessed as mandatory (required) or
// optional (with defaults) values for common types like strings, integers, booleans,
// and durations.
//
// Example usage:
//
//	cfg, err := config.New(
//		config.WithEnvPrefix("MYAPP"),
//		config.WithExtraSource(config.NewYamlConfig("config.yaml")),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Read mandatory values
//	host, err := cfg.Mandatory().String("server.host")
//	port, err := cfg.Mandatory().Int("server.port")
//
//	// Read optional values with defaults
//	timeout := cfg.Optional().Duration("server.timeout", 30*time.Second)
//
//	// Decode into a struct
//	var settings ServerSettings
//	if err := cfg.Read(&settings); err != nil {
//		log.Fatal(err)
//	}
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/txix-open/bellows"
)

// Validator defines an interface for validating configuration values.
// Implementations should return an error if the configuration structure is invalid.
type Validator interface {
	ValidateToError(value any) error
}

// Config manages application configuration from multiple sources.
// It supports loading from YAML files, environment variables, and custom sources,
// with automatic normalization of keys and value expansion.
type Config struct {
	config    map[string]string
	optional  Optional
	mandatory Mandatory

	envPrefix    string
	validator    Validator
	extraSources []Source
}

// New creates a new Config instance with the provided options.
// It loads configuration from extra sources and environment variables.
// Environment variables are filtered by the envPrefix if set.
// Keys are normalized to lowercase for case-insensitive access.
// Returns an error if any extra source fails to load.
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

// Set assigns a value to the specified key.
// The value is converted to a string representation.
func (c *Config) Set(key string, value any) {
	c.config[key] = fmt.Sprintf("%v", value)
}

// Delete removes a value by configuration key.
func (c *Config) Delete(key string) {
	delete(c.config, key)
}

// Mandatory returns a Mandatory accessor for required configuration values.
// Values accessed through Mandatory will return an error if the key is not found.
func (c *Config) Mandatory() Mandatory {
	return c.mandatory
}

// Optional returns an Optional accessor for configuration values with defaults.
// Values accessed through Optional will return the provided default if the key is not found.
func (c *Config) Optional() Optional {
	return c.optional
}

// Read decodes the configuration into the provided pointer.
// It uses mapstructure for decoding with support for type conversions (e.g., string to time.Duration).
// If a Validator is configured, it validates the decoded structure before returning.
// Returns an error if decoding or validation fails.
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

// nolint:ireturn
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

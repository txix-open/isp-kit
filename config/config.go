package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Validator interface {
	ValidateToError(value interface{}) error
}

type Config struct {
	cfg       *viper.Viper
	optional  Optional
	mandatory Mandatory

	validator Validator
	file      string
	envPrefix string
}

func New(opts ...Option) (*Config, error) {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}

	viper := viper.New()
	viper.AutomaticEnv()
	if cfg.envPrefix != "" {
		viper.SetEnvPrefix(cfg.envPrefix)
	}

	if cfg.file != "" {
		viper.SetConfigFile(cfg.file)
		err := viper.ReadInConfig()
		if err != nil {
			return nil, errors.WithMessage(err, "read config from file")
		}
	}

	cfg.cfg = viper
	cfg.optional = Optional{v: viper}
	cfg.mandatory = Mandatory{v: viper}

	return cfg, nil
}

func (c *Config) Set(key string, value interface{}) {
	c.cfg.Set(key, value)
}

func (c *Config) Mandatory() Mandatory {
	return c.mandatory
}

func (c *Config) Optional() Optional {
	return c.optional
}

func (c Config) Read(ptr interface{}) error {
	err := c.cfg.Unmarshal(&ptr)
	if err != nil {
		return errors.WithMessage(err, "unmarshal config")
	}

	if c.validator == nil {
		return nil
	}
	err = c.validator.ValidateToError(ptr)
	if err != nil {
		return errors.WithMessage(err, "validate config")
	}
	return nil
}

package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Mandatory struct {
	v *viper.Viper
}

func (m Mandatory) Int(key string) (int, error) {
	if m.v.IsSet(key) {
		return m.v.GetInt(key), nil
	}
	return 0, errors.Errorf("%s is expected in config", key)
}

func (m Mandatory) String(key string) (string, error) {
	if m.v.IsSet(key) {
		return m.v.GetString(key), nil
	}
	return "", errors.Errorf("%s is expected in config", key)
}

func (m Mandatory) Bool(key string) (bool, error) {
	if m.v.IsSet(key) {
		return m.v.GetBool(key), nil
	}
	return false, errors.Errorf("%s is expected in config", key)
}

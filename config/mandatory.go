package config

import (
	"strconv"
	"time"
)

type Mandatory struct {
	config map[string]string
}

func (m Mandatory) Int(key string) (int, error) {
	value, err := get[int](m.config, key, strconv.Atoi)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (m Mandatory) String(key string) (string, error) {
	value, err := get[string](m.config, key, func(value string) (string, error) {
		return value, nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

func (m Mandatory) Bool(key string) (bool, error) {
	value, err := get[bool](m.config, key, strconv.ParseBool)
	if err != nil {
		return false, err
	}
	return value, nil
}

func (m Mandatory) Duration(key string) (time.Duration, error) {
	value, err := get[time.Duration](m.config, key, time.ParseDuration)
	if err != nil {
		return 0, err
	}
	return value, nil
}

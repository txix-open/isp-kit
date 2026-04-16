package config

import (
	"strconv"
	"time"
)

// Mandatory provides access to required configuration values.
// All methods return an error if the requested key is not found or cannot be parsed.
// Keys are normalized to lowercase for case-insensitive access.
type Mandatory struct {
	config map[string]string
}

// Int retrieves an integer configuration value.
// Returns an error if the key is not found or the value cannot be parsed as an integer.
func (m Mandatory) Int(key string) (int, error) {
	value, err := get[int](m.config, key, strconv.Atoi)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// String retrieves a string configuration value.
// Returns an error if the key is not found.
func (m Mandatory) String(key string) (string, error) {
	value, err := get[string](m.config, key, func(value string) (string, error) {
		return value, nil
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

// Bool retrieves a boolean configuration value.
// Returns an error if the key is not found or the value cannot be parsed as a boolean.
// Accepts "1", "t", "T", "true", "TRUE", "True", "0", "f", "F", "false", "FALSE", "False".
func (m Mandatory) Bool(key string) (bool, error) {
	value, err := get[bool](m.config, key, strconv.ParseBool)
	if err != nil {
		return false, err
	}
	return value, nil
}

// Duration retrieves a time.Duration configuration value.
// Returns an error if the key is not found or the value cannot be parsed as a duration.
// Accepts formats like "30s", "1m", "1h30m" as parsed by time.ParseDuration.
func (m Mandatory) Duration(key string) (time.Duration, error) {
	value, err := get[time.Duration](m.config, key, time.ParseDuration)
	if err != nil {
		return 0, err
	}
	return value, nil
}

package config

import (
	"time"
)

// Optional provides access to configuration values with defaults.
// All methods return the provided default value if the key is not found or cannot be parsed.
// Unlike Mandatory, these methods never return errors.
type Optional struct {
	m Mandatory
}

// Int retrieves an integer configuration value or returns the default.
// Returns defValue if the key is not found or the value cannot be parsed as an integer.
func (o Optional) Int(key string, defValue int) int {
	value, err := o.m.Int(key)
	if err != nil {
		return defValue
	}
	return value
}

// String retrieves a string configuration value or returns the default.
// Returns defValue if the key is not found.
func (o Optional) String(key string, defValue string) string {
	value, err := o.m.String(key)
	if err != nil {
		return defValue
	}
	return value
}

// Bool retrieves a boolean configuration value or returns the default.
// Returns defValue if the key is not found or the value cannot be parsed as a boolean.
func (o Optional) Bool(key string, defValue bool) bool {
	value, err := o.m.Bool(key)
	if err != nil {
		return defValue
	}
	return value
}

// Duration retrieves a time.Duration configuration value or returns the default.
// Returns defValue if the key is not found or the value cannot be parsed as a duration.
func (o Optional) Duration(key string, defValue time.Duration) time.Duration {
	value, err := o.m.Duration(key)
	if err != nil {
		return defValue
	}
	return value
}

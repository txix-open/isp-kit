package config

import (
	"time"
)

type Optional struct {
	m Mandatory
}

func (o Optional) Int(key string, defValue int) int {
	value, err := o.m.Int(key)
	if err != nil {
		return defValue
	}
	return value
}

func (o Optional) String(key string, defValue string) string {
	value, err := o.m.String(key)
	if err != nil {
		return defValue
	}
	return value
}

func (o Optional) Bool(key string, defValue bool) bool {
	value, err := o.m.Bool(key)
	if err != nil {
		return defValue
	}
	return value
}

func (o Optional) Duration(key string, defValue time.Duration) time.Duration {
	value, err := o.m.Duration(key)
	if err != nil {
		return defValue
	}
	return value
}

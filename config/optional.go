package config

import (
	"github.com/spf13/viper"
)

type Optional struct {
	v *viper.Viper
}

func (o Optional) Int(key string, defValue int) int {
	if o.v.IsSet(key) {
		return o.v.GetInt(key)
	}
	return defValue
}

func (o Optional) String(key string, defValue string) string {
	if o.v.IsSet(key) {
		return o.v.GetString(key)
	}
	return defValue
}

func (o Optional) Bool(key string, defValue bool) bool {
	if o.v.IsSet(key) {
		return o.v.GetBool(key)
	}
	return defValue
}

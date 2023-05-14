package config_tv

import (
	"github.com/mitchellh/mapstructure"
)

type PluginConfig map[string]interface{}

func ConvertPluginConfig[T any](pluginConfig PluginConfig) *T {
	var cfgObj = new(T)
	if err := mapstructure.Decode(pluginConfig, cfgObj); err != nil {
		panic(err)
	}
	return cfgObj
}

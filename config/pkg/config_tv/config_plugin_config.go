package config_tv

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os"
)

type ConfigPluginConfig struct {
	Plugin map[string]PluginConfig
}

func GetConfigPluginConfigFromEnv() *ConfigPluginConfig {
	var tv = TypeAndValue{
		Type:  os.Getenv("CONFIGPLUGINCONFIG_TYPE"),
		Value: os.Getenv("CONFIGPLUGINCONFIG_VALUE"),
	}
	var configPluginConfigRawData = tv.ReadRawDataNoPlugin()
	var configConfigLoader = viper.New()
	configConfigLoader.SetConfigType("toml")
	if err := configConfigLoader.MergeConfig(bytes.NewReader(configPluginConfigRawData)); err != nil {
		panic(errors.Wrapf(err, "read config plugin config err from tv:%s", tv))
	}
	var configPluginConfig = new(ConfigPluginConfig)
	if err := configConfigLoader.Unmarshal(configPluginConfig); err != nil {
		panic(errors.Wrap(err, "err unmarshal config"))
	}
	return configPluginConfig
}

func GetAndUnmarshalMainConfigFromEnv(mainConfig interface{}, pluginMap TypePluginMap) {
	var tv = TypeAndValue{
		Type:  os.Getenv("MAINCONFIG_TYPE"),
		Value: os.Getenv("MAINCONFIG_VALUE"),
	}
	var mainConfigRawData = tv.ReadRawData(pluginMap)
	var mainConfigLoader = viper.New()
	mainConfigLoader.SetConfigType("toml")
	if err := mainConfigLoader.MergeConfig(bytes.NewReader(mainConfigRawData)); err != nil {
		panic(errors.Wrapf(err, "read main config err from tv:%s", tv))
	}
	if err := mainConfigLoader.Unmarshal(mainConfig); err != nil {
		panic(errors.Wrap(err, "err unmarshal config"))
	}
}

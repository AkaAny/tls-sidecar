package config_tv

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

const (
	TypePath  = "path"
	TypeEmbed = "embed"
)

type TypePluginMap map[string]TypeAndValuePlugin

func (x TypePluginMap) AddPlugin(pluginName string, plugin TypeAndValuePlugin) {
	x[pluginName] = plugin
}

type TypeAndValue struct {
	Type  string
	Value string
}

func (x TypeAndValue) String() string {
	rawData, _ := json.Marshal(x)
	return string(rawData)
}

func (x TypeAndValue) ReadRawDataNoPlugin() []byte {
	switch x.Type {
	case TypePath:
		rawData, err := os.ReadFile(x.Value)
		if err != nil {
			panic(errors.Wrapf(err, "err read from path:%s", x.Value))
		}
		return rawData
	case TypeEmbed:
		return []byte(x.Value)
	default:
		panic(errors.Errorf("unsupported type:%s", x.Type))
	}
}

func (x TypeAndValue) ReadRawData(typePluginMap TypePluginMap) []byte {
	switch x.Type {
	case TypePath:
		return x.ReadRawDataNoPlugin()
	case TypeEmbed:
		return x.ReadRawDataNoPlugin()
	default:
		plugin, ok := typePluginMap[x.Type]
		if !ok {
			panic(errors.Errorf("unsupported type:%s", x.Type))
		}
		var rawData = plugin.ReadRawData(x)
		return rawData
	}
}

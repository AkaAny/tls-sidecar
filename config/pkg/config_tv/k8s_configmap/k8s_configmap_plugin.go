package k8s_configmap

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"tls-sidecar/config/pkg/config_tv"
)

const PluginName = "k8s_configmap"

func NewK8sConfigPluginFromConfig(pluginConfigMap config_tv.PluginConfig) *K8SConfigMapPlugin {
	var baseK8sPlugin = config_tv.NewBaseK8sPluginFromConfig(pluginConfigMap)
	return &K8SConfigMapPlugin{baseK8sPlugin}
}

type K8SConfigMapPlugin struct {
	*config_tv.BaseK8sPlugin
}

func (k *K8SConfigMapPlugin) ReadRawData(tv config_tv.TypeAndValue) []byte {
	//ns,secretName,secretKey:=tv.Extra["namespace"].(string),tv.Extra["name"].(string),tv.Extra["key"].(string)
	var pathParts = strings.Split(tv.Value, "/")
	ns, configMapName, configMapKey := pathParts[0], pathParts[1], pathParts[2]
	configMapInfo, err := k.ClientSet.CoreV1().
		ConfigMaps(ns).
		Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	var valueStr = configMapInfo.Data[configMapKey]
	return []byte(valueStr)
}

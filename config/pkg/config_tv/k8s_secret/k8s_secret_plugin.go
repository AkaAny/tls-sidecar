package k8s_secret

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"tls-sidecar/config/pkg/config_tv"
)

const PluginName = "k8s_secret"

type K8sSecretPluginConfig struct {
	Mode       string                 `mapstructure:"mode"`
	KubeConfig config_tv.TypeAndValue `mapstructure:"kubeConfig"`
}

func NewK8sSecretPluginFromConfig(pluginConfigMap config_tv.PluginConfig) *K8SSecretPlugin {
	var baseK8sPlugin = config_tv.NewBaseK8sPluginFromConfig(pluginConfigMap)
	return &K8SSecretPlugin{baseK8sPlugin}
}

type K8SSecretPlugin struct {
	*config_tv.BaseK8sPlugin
}

func (k *K8SSecretPlugin) ReadRawData(tv config_tv.TypeAndValue) []byte {
	//ns,secretName,secretKey:=tv.Extra["namespace"].(string),tv.Extra["name"].(string),tv.Extra["key"].(string)
	var pathParts = strings.Split(tv.Value, "/")
	ns, secretName, secretKey := pathParts[0], pathParts[1], pathParts[2]
	secretInfo, err := k.ClientSet.CoreV1().
		Secrets(ns).
		Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		panic(err)
	}
	return secretInfo.Data[secretKey]
}

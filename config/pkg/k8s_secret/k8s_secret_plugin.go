package k8s_secret

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"tls-sidecar/config/pkg"
)

const PluginName = "k8s_secret"

type K8sSecretPluginConfig struct {
	Mode       string           `mapstructure:"mode"`
	KubeConfig pkg.TypeAndValue `mapstructure:"kubeConfig"`
}

func NewK8sSecretPluginFromConfig(pluginConfigMap pkg.PluginConfig) *K8SSecretPlugin {
	var baseK8sPlugin = pkg.NewBaseK8sPluginFromConfig(pluginConfigMap)
	return &K8SSecretPlugin{baseK8sPlugin}
}

type K8SSecretPlugin struct {
	*pkg.BaseK8sPlugin
}

func (k *K8SSecretPlugin) ReadRawData(tv pkg.TypeAndValue) []byte {
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

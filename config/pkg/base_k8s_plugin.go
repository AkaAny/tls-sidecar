package pkg

import (
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type BaseK8sPlugin struct {
	ClientSet *kubernetes.Clientset
}

type BaseK8sPluginConfig struct {
	Mode       string       `mapstructure:"mode"`
	KubeConfig TypeAndValue `mapstructure:"kubeConfig"`
}

func NewBaseK8sPlugin(cfg *rest.Config) *BaseK8sPlugin {
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		panic(err)
	}
	return &BaseK8sPlugin{ClientSet: clientSet}
}

func NewBaseK8sPluginFromConfig(pluginConfigMap PluginConfig) *BaseK8sPlugin {
	var pluginConfig = ConvertPluginConfig[BaseK8sPluginConfig](pluginConfigMap)
	switch pluginConfig.Mode {
	case "kubeconfig":
		var kubeConfigData = pluginConfig.KubeConfig.ReadRawDataNoPlugin()
		clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeConfigData)
		if err != nil {
			panic(err)
		}
		restConfig, err := clientConfig.ClientConfig()
		if err != nil {
			panic(err)
		}
		var plugin = NewBaseK8sPlugin(restConfig)
		return plugin
	case "inCluster":
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}
		var plugin = NewBaseK8sPlugin(restConfig)
		return plugin
	default:
		panic(errors.Errorf("unknown mode:%s", pluginConfig.Mode))
	}
}

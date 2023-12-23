package main

import (
	"crypto/x509"
	"fmt"
	config_tv "github.com/AkaAny/config-tv"
	"github.com/AkaAny/config-tv/plugin/k8s_configmap"
	"github.com/AkaAny/config-tv/plugin/k8s_secret"
	tls_sidecar "github.com/AkaAny/tls-sidecar"
	"github.com/AkaAny/tls-sidecar/config"
	tls_on_http "github.com/AkaAny/tls-sidecar/http_handler"
	"github.com/AkaAny/tls-sidecar/ws_handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"os"
	"os/signal"
)

func main() {
	fmt.Println("sidecar starts working")
	var configPluginConfig = config_tv.GetConfigPluginConfigFromEnv()
	var pluginMap = make(config_tv.TypePluginMap)
	{
		var pluginConfigMap = configPluginConfig.Plugin[k8s_secret.PluginName]
		var k8sSecretTypeKVPlugin = k8s_secret.NewK8sSecretPluginFromConfig(pluginConfigMap)
		pluginMap[k8s_secret.PluginName] = k8sSecretTypeKVPlugin
	}
	{
		var pluginConfigMap = configPluginConfig.Plugin[k8s_configmap.PluginName]
		var k8sConfigMapTypeKVPlugin = k8s_configmap.NewK8sConfigPluginFromConfig(pluginConfigMap)
		pluginMap[k8s_configmap.PluginName] = k8sConfigMapTypeKVPlugin
	}
	var mainConfig = new(config.SidecarConfig)
	{
		config_tv.GetAndUnmarshalMainConfigFromEnv(mainConfig, pluginMap)
	}
	var trustedDeployCerts = lo.Map(mainConfig.RPC.Inbound.TrustedDeployCertificates,
		func(item *config.CertificateTypeAndValue, index int) *x509.Certificate {
			return item.ReadAndParse(pluginMap)
		})
	var serviceCert = mainConfig.RPC.BackendServiceCertificate.ReadAndParse(pluginMap)
	var serviceKey = mainConfig.RPC.BackendServiceKey.ReadAndParse(pluginMap)
	var engine = gin.Default()
	var corsMiddleware = func() gin.HandlerFunc {
		var defaultConfig = cors.DefaultConfig()
		defaultConfig.AllowOriginFunc = func(origin string) bool {
			return true
		}
		defaultConfig.AllowAllOrigins = false
		defaultConfig.AllowWebSockets = true
		defaultConfig.AddAllowHeaders("X-Connection-ID")
		defaultConfig.AddExposeHeaders("X-Connection-ID")
		return cors.New(defaultConfig)
	}()
	engine.Use(corsMiddleware)
	var serverTLSConfigFactory = &tls_sidecar.ServerTLSConfigFactory{
		SelfKey:          serviceKey,
		SelfCert:         serviceCert,
		TrustDeployCerts: trustedDeployCerts,
	}
	var clientTLSConfigFactory = &tls_sidecar.ClientTLSConfigFactory{
		SelfKey:    serviceKey,
		SelfCert:   serviceCert,
		ParentCert: trustedDeployCerts[0],
	}
	var websocketSidecarServer = &ws_handler.WebSocketSidecarServer{
		ServerTLSConfigFactory: serverTLSConfigFactory,
		ClientTLSConfigFactory: clientTLSConfigFactory,
		DeployHostMap:          mainConfig.RPC.Outbound.DeployIDHostMap,
		BackendServiceHost:     mainConfig.RPC.Inbound.BackendServiceHost,
	}
	var httpSidecarServer = &tls_on_http.HttpSidecarServer{
		ServerTLSConfigFactory: serverTLSConfigFactory,
		ClientTLSConfigFactory: clientTLSConfigFactory,
		BackendServiceHost:     mainConfig.RPC.Inbound.BackendServiceHost,
		DeployIDHostMap:        mainConfig.RPC.Outbound.DeployIDHostMap,
	}
	{
		var wsGroup = engine.Group("/")
		websocketSidecarServer.InitRequestRouter(wsGroup) //默认用ws协议，除非特别指定
		var httpGroup = engine.Group("/http")
		httpSidecarServer.InitRequestRouter(httpGroup)
	}
	go func() {
		if err := engine.Run(":9090"); err != nil {
			panic(err)
		}
	}()
	var wrapEngine = gin.Default()
	wrapEngine.Use(corsMiddleware)
	{
		websocketSidecarServer.InitWrapRouter(wrapEngine)
		var httpGroup = wrapEngine.Group("/http")
		httpSidecarServer.InitWrapRouter(httpGroup)
	}
	go func() {
		if err := wrapEngine.Run(":18080"); err != nil {
			panic(err)
		}
	}()
	//inbound.Main(trustedDeployCerts,
	//	serviceCert, serviceKey,
	//	mainConfig.RPC.Inbound.BackendServiceHost)
	//var selfDeployCert = mainConfig.RPC.Outbound.SelfDeployCertificate.ReadAndParse()
	//outbound.Main(selfDeployCert,
	//	serviceCert, serviceKey,
	//	mainConfig.RPC.Outbound.DeployIDHostMap)
	var sigChan = make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	//fmt.Println("interrupted")
}

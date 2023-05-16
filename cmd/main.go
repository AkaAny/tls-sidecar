package main

import (
	"crypto/x509"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"io"
	"net/http"
	"os"
	"os/signal"
	tls_sidecar "tls-sidecar"
	"tls-sidecar/config"
	"tls-sidecar/config/pkg/config_tv"
	"tls-sidecar/config/pkg/config_tv/k8s_configmap"
	"tls-sidecar/config/pkg/config_tv/k8s_secret"
	"tls-sidecar/trust_center"
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
	var wsHandler = tls_sidecar.WSHandler{
		ServiceKey:         serviceKey,
		ServiceCert:        serviceCert,
		TrustDeployCerts:   trustedDeployCerts,
		BackendServiceHost: mainConfig.RPC.Inbound.BackendServiceHost,
	}
	var engine = gin.Default()
	var defaultConfig = cors.DefaultConfig()
	defaultConfig.AllowOriginFunc = func(origin string) bool {
		fmt.Println("cors origin:", origin)
		return true
	}
	defaultConfig.AllowWebSockets = true
	engine.Use(cors.New(defaultConfig))
	engine.GET("/tlsRequest", func(c *gin.Context) {
		wsHandler.Attach(c.Writer, c.Request)
	})
	go func() {
		if err := engine.Run(":9090"); err != nil {
			panic(err)
		}
	}()
	var wrapEngine = gin.Default()
	wrapEngine.POST("/wrap", func(c *gin.Context) {
		var targetMethod = c.GetHeader("X-Target-Method")
		var targetPath = c.GetHeader("X-Target-Path")
		var targetQuery = c.GetHeader("X-Target-Query")
		var targetDeployID = c.GetHeader("X-Target-Deploy")
		var targetServiceID = c.GetHeader("X-Target-Service")
		deployHost, ok := mainConfig.RPC.Outbound.DeployIDHostMap[targetDeployID]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"msg": errors.Errorf("deploy id:%s does not exist", targetDeployID).Error(),
			})
			return
		}
		var targetWSUrl = fmt.Sprintf("http://%s/%s/tlsRequest", deployHost, targetServiceID)
		var realUrl = fmt.Sprintf("http://%s%s?%s", targetServiceID, targetPath, targetQuery)
		req, err := http.NewRequest(targetMethod, realUrl, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": errors.Wrap(err, "new http request").Error(),
			})
			return
		}
		req.Header = c.Request.Header
		req.Header.Set(trust_center.IdentityTypeHeaderKey, trust_center.IdentityTypeCertRPC)
		resp, err := tls_sidecar.NewWSClient(tls_sidecar.WSClientParam{
			TargetWSURL: targetWSUrl,
			SelfKey:     serviceKey,
			SelfCert:    serviceCert,
			ParentCert:  trustedDeployCerts[0],
		}, req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": errors.Wrap(err, "do tls request").Error(),
			})
			return
		}
		for headerKey, headerValues := range resp.Header {
			c.Writer.Header()[headerKey] = headerValues
		}
		defer resp.Body.Close()
		respBodyData, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": errors.Wrap(err, "read response body"),
			})
			return
		}
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBodyData)
		//if _, err := c.Writer.Write(respBodyData); err != nil {
		//	c.JSON(http.StatusInternalServerError, gin.H{
		//		"msg": errors.Wrap(err, "write response body"),
		//	})
		//	return
		//}
	})
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

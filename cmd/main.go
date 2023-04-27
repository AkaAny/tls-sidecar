package main

import (
	"crypto/x509"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	tls_sidecar "tls-sidecar"
	"tls-sidecar/config"
)

func main() {
	fmt.Println("sidecar starts working")
	var mainConfigLoader = viper.New()
	var sidecarConfigPath = os.Getenv("SIDECAR_CONFIG_PATH")
	{
		mainConfigLoader.SetConfigFile(sidecarConfigPath)
	}
	if err := mainConfigLoader.ReadInConfig(); err != nil {
		panic(errors.Wrapf(err, "read config err from path:%s", sidecarConfigPath))
	}
	var sidecarConfig = new(config.SidecarConfig)
	if err := mainConfigLoader.Unmarshal(sidecarConfig); err != nil {
		panic(errors.Wrap(err, "err unmarshal config"))
	}
	var trustedDeployCerts = lo.Map(sidecarConfig.RPC.Inbound.TrustedDeployCertificates,
		func(item *config.CertificateTypeAndValue, index int) *x509.Certificate {
			return item.ReadAndParse()
		})
	var serviceCert = sidecarConfig.RPC.BackendServiceCertificate.ReadAndParse()
	var serviceKey = sidecarConfig.RPC.BackendServiceKey.ReadAndParse()
	var wsHandler = tls_sidecar.WSHandler{
		ServiceKey:       serviceKey,
		ServiceCert:      serviceCert,
		TrustDeployCerts: trustedDeployCerts,
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

	engine.POST("/wrap", func(c *gin.Context) {
		var targetMethod = c.GetHeader("X-Target-Method")
		var targetPath = c.GetHeader("X-Target-Path")
		var targetDeployID = c.GetHeader("X-Target-Deploy")
		var targetServiceID = c.GetHeader("X-Target-Service")
		deployHost, ok := sidecarConfig.RPC.Outbound.DeployIDHostMap[targetDeployID]
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"msg": errors.Errorf("deploy id:%s does not exist", targetDeployID).Error(),
			})
			return
		}
		var targetWSUrl = fmt.Sprintf("http://%s/%s/tlsRequest", deployHost, targetServiceID)
		var realUrl = fmt.Sprintf("http://%s%s", targetServiceID, targetPath)
		req, err := http.NewRequest(targetMethod, realUrl, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"msg": errors.Wrap(err, "new http request").Error(),
			})
			return
		}
		resp, err := tls_sidecar.NewWSClient(tls_sidecar.WSClientParam{
			TargetWSURL: targetWSUrl,
			SelfKey:     serviceKey,
			SelfCert:    serviceCert,
			DeployCert:  trustedDeployCerts[0],
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
	if err := engine.Run(":9090"); err != nil {
		panic(err)
	}
	//inbound.Main(trustedDeployCerts,
	//	serviceCert, serviceKey,
	//	sidecarConfig.RPC.Inbound.BackendServiceHost)
	//var selfDeployCert = sidecarConfig.RPC.Outbound.SelfDeployCertificate.ReadAndParse()
	//outbound.Main(selfDeployCert,
	//	serviceCert, serviceKey,
	//	sidecarConfig.RPC.Outbound.DeployIDHostMap)
	//var sigChan = make(chan os.Signal)
	//signal.Notify(sigChan, os.Interrupt)
	//<-sigChan
	//fmt.Println("interrupted")
}

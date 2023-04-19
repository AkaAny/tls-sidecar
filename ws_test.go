package tls_sidecar

import (
	"crypto/x509"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"testing"
	"tls-sidecar/cert_manager"
)

func TestWSHandler_Attach(t *testing.T) {
	var deployHDUCert = cert_manager.ParseX509CertificateFromFile("deploy-hdu.crt")
	var serviceKey = cert_manager.ParsePKCS8PEMPrivateKeyFromFile("rpc-service-storage.key")
	var serviceCert = cert_manager.ParseX509CertificateFromFile("rpc-service-storage.crt")
	var handler = &WSHandler{
		ServiceKey:       serviceKey,
		ServiceCert:      serviceCert,
		TrustDeployCerts: []*x509.Certificate{deployHDUCert},
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
		handler.Attach(c.Writer, c.Request)
	})
	if err := engine.Run(":9090"); err != nil {
		panic(err)
	}
}

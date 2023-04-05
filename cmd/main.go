package main

import (
	"crypto/x509"
	"fmt"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"tls-sidecar/cmd/inbound"
	"tls-sidecar/cmd/outbound"
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
	inbound.Main(trustedDeployCerts,
		serviceCert, serviceKey,
		sidecarConfig.RPC.Inbound.BackendServiceHost)
	var selfDeployCert = sidecarConfig.RPC.Outbound.SelfDeployCertificate.ReadAndParse()
	outbound.Main(selfDeployCert,
		serviceCert, serviceKey,
		sidecarConfig.RPC.Outbound.DeployIDHostMap)
	var sigChan = make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("interrupted")
}

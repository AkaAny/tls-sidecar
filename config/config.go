package config

import (
	"github.com/AkaAny/config-tv"
)

type SidecarConfig struct {
	Config *config_tv.ConfigPluginConfig
	RPC    *RPCConfig `mapstructure:"rpc"`
}

type RPCConfig struct {
	BackendServiceKey         *RSAPrivateKeyTypeAndValue
	BackendServiceCertificate *CertificateTypeAndValue
	Inbound                   *InboundConfig
	Outbound                  *OutboundConfig
}

type InboundConfig struct {
	TrustedDeployCertificates []*CertificateTypeAndValue
	BackendServiceHost        string
}

type OutboundConfig struct {
	SelfDeployCertificate *CertificateTypeAndValue
	DeployIDHostMap       map[string]string `mapstructure:"deployIDHostMap"`
}

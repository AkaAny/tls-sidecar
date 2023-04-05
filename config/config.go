package config

type SidecarConfig struct {
	RPC *RPCConfig `mapstructure:"rpc"`
}

type RPCConfig struct {
	RPCCACertificate *CertificateTypeAndValue `mapstructure:"rpcCACertificate"`

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

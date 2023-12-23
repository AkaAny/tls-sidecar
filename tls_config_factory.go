package tls_sidecar

import (
	"crypto/rsa"
	tls2 "crypto/tls"
	"crypto/x509"
	"github.com/samber/lo"
)

type ServerTLSConfigFactory struct {
	SelfKey          *rsa.PrivateKey
	SelfCert         *x509.Certificate
	TrustDeployCerts []*x509.Certificate
}

func (tcf ServerTLSConfigFactory) NewServerConfig() *tls2.Config {
	var tlsCert = NewTLSCertificate(tcf.SelfKey, tcf.SelfCert)

	var caPool = x509.NewCertPool()
	lo.ForEach(tcf.TrustDeployCerts, func(item *x509.Certificate, index int) {
		caPool.AddCert(item)
	})
	//serviceCertInfo, err := trust_center.NewServiceCertificate(h.ServiceCert)
	//if err != nil {
	//	panic(errors.Wrap(err, "invalid service certificate"))
	//}
	var tlsConfig = &tls2.Config{
		Certificates:          []tls2.Certificate{tlsCert},
		GetCertificate:        nil,
		GetClientCertificate:  nil,
		GetConfigForClient:    nil,
		VerifyPeerCertificate: nil,
		VerifyConnection:      nil,
		//RootCAs:               caPool,
		ServerName:         "",
		ClientAuth:         tls2.RequireAndVerifyClientCert, //tls2.RequireAndVerifyClientCert
		ClientCAs:          caPool,
		InsecureSkipVerify: false,
	}
	return tlsConfig
}

type ClientTLSConfigFactory struct {
	SelfKey    *rsa.PrivateKey
	SelfCert   *x509.Certificate
	ParentCert *x509.Certificate //对于服务是deploy，对于用户是staff ca
}

func (tsc *ClientTLSConfigFactory) NewClientConfig() *tls2.Config {
	var tlsCert = NewTLSCertificate(tsc.SelfKey, tsc.SelfCert)

	var caPool = x509.NewCertPool()
	//caPool.AddCert(rootCACert)
	caPool.AddCert(tsc.ParentCert)

	var tlsConfig = &tls2.Config{
		Certificates: []tls2.Certificate{
			tlsCert,
		},
		RootCAs: caPool,

		VerifyPeerCertificate: nil,
		VerifyConnection:      nil,

		ServerName:         "",
		InsecureSkipVerify: true,
	}
	return tlsConfig
}

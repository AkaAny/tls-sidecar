package tls_sidecar

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"github.com/samber/lo"
)

func NewTLSCertificate(serviceKey *rsa.PrivateKey, serviceCerts ...*x509.Certificate) tls.Certificate {
	var certDers = lo.Map(serviceCerts, func(item *x509.Certificate, index int) []byte {
		return item.Raw
	})
	var tlsCert = tls.Certificate{
		Certificate: certDers,
		PrivateKey:  serviceKey,
	}
	return tlsCert
}

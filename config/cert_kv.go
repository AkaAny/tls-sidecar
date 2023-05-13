package config

import (
	"crypto/rsa"
	"crypto/x509"
	"github.com/pkg/errors"
	"tls-sidecar/cert_manager"
	"tls-sidecar/config/pkg"
)

const (
	TypePath  = "path"
	TypeEmbed = "embed"
)

type CertificateTypeAndValue pkg.TypeAndValue

func (x CertificateTypeAndValue) ReadAndParse(pluginMap pkg.TypePluginMap) *x509.Certificate {
	var super = pkg.TypeAndValue(x)
	var rawData = super.ReadRawData(pluginMap)
	var der = cert_manager.ReadSingleDerFromPEMData(rawData)
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		panic(errors.Wrap(err, "err parse x509 certificate from der"))
	}
	return cert
}

type RSAPrivateKeyTypeAndValue pkg.TypeAndValue

func (x RSAPrivateKeyTypeAndValue) ReadAndParse(pluginMap pkg.TypePluginMap) *rsa.PrivateKey {
	var super = pkg.TypeAndValue(x)
	var rawData = super.ReadRawData(pluginMap)
	var der = cert_manager.ReadSingleDerFromPEMData(rawData)
	privateKeyObj, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		panic(errors.Wrap(err, "err parse pkcs1 rsa private key from der"))
	}
	return privateKeyObj.(*rsa.PrivateKey)
}

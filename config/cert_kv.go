package config

import (
	"crypto/rsa"
	"crypto/x509"
	"github.com/pkg/errors"
	"tls-sidecar/cert_manager"
	"tls-sidecar/config/pkg/config_tv"
)

const (
	TypePath  = "path"
	TypeEmbed = "embed"
)

type CertificateTypeAndValue config_tv.TypeAndValue

func (x CertificateTypeAndValue) ReadAndParse(pluginMap config_tv.TypePluginMap) *x509.Certificate {
	var super = config_tv.TypeAndValue(x)
	var rawData = super.ReadRawData(pluginMap)
	var der = cert_manager.ReadSingleDerFromPEMData(rawData)
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		panic(errors.Wrap(err, "err parse x509 certificate from der"))
	}
	return cert
}

type RSAPrivateKeyTypeAndValue config_tv.TypeAndValue

func (x RSAPrivateKeyTypeAndValue) ReadAndParse(pluginMap config_tv.TypePluginMap) *rsa.PrivateKey {
	var super = config_tv.TypeAndValue(x)
	var rawData = super.ReadRawData(pluginMap)
	var der = cert_manager.ReadSingleDerFromPEMData(rawData)
	privateKeyObj, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		panic(errors.Wrap(err, "err parse pkcs1 rsa private key from der"))
	}
	return privateKeyObj.(*rsa.PrivateKey)
}

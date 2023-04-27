package config

import (
	"crypto/rsa"
	"crypto/x509"
	"github.com/pkg/errors"
	"os"
	"tls-sidecar/cert_manager"
)

const (
	TypePath  = "path"
	TypeEmbed = "embed"
)

type TypeAndValue struct {
	Type  string
	Value string
}

func (x TypeAndValue) ReadRawData() []byte {
	switch x.Type {
	case TypePath:
		rawData, err := os.ReadFile(x.Value)
		if err != nil {
			panic(errors.Wrapf(err, "err read from path:%s", x.Value))
		}
		return rawData
	case TypeEmbed:
		return []byte(x.Value)
	default:
		panic(errors.Errorf("unsupported type:%s", x.Type))
	}
}

type CertificateTypeAndValue TypeAndValue

func (x CertificateTypeAndValue) ReadAndParse() *x509.Certificate {
	var super = TypeAndValue(x)
	var rawData = super.ReadRawData()
	var der = cert_manager.ReadSingleDerFromPEMData(rawData)
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		panic(errors.Wrap(err, "err parse x509 certificate from der"))
	}
	return cert
}

type RSAPrivateKeyTypeAndValue TypeAndValue

func (x RSAPrivateKeyTypeAndValue) ReadAndParse() *rsa.PrivateKey {
	var super = TypeAndValue(x)
	var rawData = super.ReadRawData()
	var der = cert_manager.ReadSingleDerFromPEMData(rawData)
	privateKeyObj, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		panic(errors.Wrap(err, "err parse pkcs1 rsa private key from der"))
	}
	return privateKeyObj.(*rsa.PrivateKey)
}

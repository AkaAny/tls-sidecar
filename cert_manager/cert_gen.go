package cert_manager

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"github.com/pkg/errors"
	"math/big"
	mathRand "math/rand"
	"time"
)

func GenerateTLSCertificate(subject pkix.Name, willHaveChildren bool,
	parentCert *x509.Certificate, parentKey *rsa.PrivateKey,
	dnsNames []string) (certDer []byte, privateKey *rsa.PrivateKey, err error) {
	var issuer = pkix.Name{}
	if parentCert != nil {
		issuer = parentCert.Subject
	}
	var cert = &x509.Certificate{
		SerialNumber:          big.NewInt(mathRand.Int63()),
		Issuer:                issuer,
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              0,
		Extensions:            nil,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  parentCert == nil || willHaveChildren,

		DNSNames:       dnsNames,
		EmailAddresses: nil,
		IPAddresses:    nil,
		URIs:           nil,
	}
	privateKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, errors.Wrap(err, "generate RSA key")
	}
	if parentCert == nil && parentKey == nil { //CA自签
		parentCert = cert
		parentKey = privateKey
	}
	certDer, err = x509.CreateCertificate(rand.Reader, cert, parentCert, &privateKey.PublicKey, parentKey)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create certificate")
	}
	return certDer, privateKey, nil
}

package cert_manager

import (
	"crypto/x509/pkix"
	"os"
	"testing"
)

func TestGenerateDeployCACertificate(t *testing.T) {
	parentKey := ParsePEMPrivateKeyFromFile("../rpc-ca.key")
	parentCert := ParseX509CertificateFromFile("../rpc-ca.crt")
	certDer, privateKey, err := GenerateTLSCertificate(pkix.Name{
		Organization:       []string{"JianLi"},
		OrganizationalUnit: []string{"JianLi Deployment"},
		CommonName:         "hdu",
	}, true, parentCert, parentKey, []string{"localhost"})
	if err != nil {
		panic(err)
	}
	var pemEncodedPrivateKey = MarshalRSAPrivateKeyUsingPEM(privateKey)
	os.WriteFile("../deploy-hdu.key", pemEncodedPrivateKey, os.ModePerm)
	var pemEncodedCert = MarshalCertificateDerUsingPEM(certDer)
	os.WriteFile("../deploy-hdu.crt", pemEncodedCert, os.ModePerm)

}

func TestGenerateServiceARPCCertificate(t *testing.T) {
	parentKey := ParsePEMPrivateKeyFromFile("../deploy-hdu.key")
	parentCert := ParseX509CertificateFromFile("../deploy-hdu.crt")
	certDer, privateKey, err := GenerateTLSCertificate(pkix.Name{
		Organization:       []string{"JianLi"},
		OrganizationalUnit: []string{"hdu"},
		CommonName:         "company",
	}, false, parentCert, parentKey, []string{"localhost"})
	if err != nil {
		panic(err)
	}
	var pemEncodedPrivateKey = MarshalRSAPrivateKeyUsingPEM(privateKey)
	os.WriteFile("../rpc-service-company.key", pemEncodedPrivateKey, os.ModePerm)
	var pemEncodedCert = MarshalCertificateDerUsingPEM(certDer)
	os.WriteFile("../rpc-service-company.crt", pemEncodedCert, os.ModePerm)
}

func TestGenerateRPCCACertificate(t *testing.T) {
	parentKey := ParsePKCS8PEMPrivateKeyFromFile("../jl-root-ca.key")
	parentCert := ParseX509CertificateFromFile("../jl-root-ca.crt")
	certDer, privateKey, err := GenerateTLSCertificate(pkix.Name{
		Organization:       []string{"JianLi"},
		OrganizationalUnit: []string{"JianLi RPC"},
		CommonName:         "RPC Certificate Authority",
	}, true, parentCert, parentKey, nil)
	if err != nil {
		panic(err)
	}
	var pemEncodedPrivateKey = MarshalRSAPrivateKeyUsingPEM(privateKey)
	os.WriteFile("../rpc-ca.key", pemEncodedPrivateKey, os.ModePerm)
	var pemEncodedCert = MarshalCertificateDerUsingPEM(certDer)
	os.WriteFile("../rpc-ca.crt", pemEncodedCert, os.ModePerm)
}

package cert_manager

import (
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"os"
	"testing"
)

var defaultDnsNames = []string{"jldev.hdu.edu.cn", "jl.hdu.edu.cn"}

func TestGenerateDeployCACertificate(t *testing.T) {
	var deployID = "center"
	parentKey := ParsePKCS8PEMPrivateKeyFromFile("../jl-root-ca.key")
	parentCert := ParseX509CertificateFromFile("../jl-root-ca.crt")
	certDer, privateKey, err := GenerateTLSCertificate(pkix.Name{
		Organization:       []string{"JianLi"},
		OrganizationalUnit: []string{"JianLi Deployment"},
		CommonName:         deployID,
		ExtraNames: []pkix.AttributeTypeAndValue{
			{Type: DeployIDOID, Value: deployID},
		},
	}, true, parentCert, parentKey, []string{"jldev.hdu.edu.cn", "jl.hdu.edu.cn"})
	if err != nil {
		panic(err)
	}
	var pemEncodedPrivateKey = MarshalRSAPrivateKeyUsingPKCS8PEM(privateKey)
	os.WriteFile(fmt.Sprintf("../deploy-%s.key", deployID), pemEncodedPrivateKey, os.ModePerm)
	var pemEncodedCert = MarshalCertificateDerUsingPEM(certDer)
	os.WriteFile(fmt.Sprintf("../deploy-%s.crt", deployID), pemEncodedCert, os.ModePerm)

}

var (
	CertificateTypeOID = asn1.ObjectIdentifier{1, 1, 106108, 0, 1, 1}
	DeployIDOID        = asn1.ObjectIdentifier{1, 1, 106108, 1, 1, 1}
	ServiceIDOID       = asn1.ObjectIdentifier{1, 1, 106108, 2, 1, 1}
	UserIDOID          = asn1.ObjectIdentifier{1, 1, 106108, 3, 1, 1}
	UserTypeOID        = asn1.ObjectIdentifier{1, 1, 106108, 3, 1, 2}
)

func TestGenerateServiceARPCCertificate(t *testing.T) {
	var deployID = "center"
	var serviceID = "supportive-data"
	parentKey := ParsePKCS8PEMPrivateKeyFromFile(fmt.Sprintf("../deploy-%s.key", deployID))
	parentCert := ParseX509CertificateFromFile(fmt.Sprintf("../deploy-%s.crt", deployID))
	certDer, privateKey, err := GenerateTLSCertificate(pkix.Name{
		Organization:       []string{"JianLi"},
		OrganizationalUnit: []string{deployID},
		CommonName:         serviceID,
		ExtraNames: []pkix.AttributeTypeAndValue{
			{Type: CertificateTypeOID, Value: "service"},
			{Type: ServiceIDOID, Value: serviceID},
		},
	}, false, parentCert, parentKey, []string{"jldev.hdu.edu.cn", "jl.hdu.edu.cn"})
	if err != nil {
		panic(err)
	}
	var pemEncodedPrivateKey = MarshalRSAPrivateKeyUsingPKCS8PEM(privateKey)
	os.WriteFile(fmt.Sprintf("../rpc-service-%s.key", serviceID), pemEncodedPrivateKey, os.ModePerm)
	var pemEncodedCert = MarshalCertificateDerUsingPEM(certDer)
	os.WriteFile(fmt.Sprintf("../rpc-service-%s.crt", serviceID), pemEncodedCert, os.ModePerm)
}

func TestGenerateDeployStaffCACertificate(t *testing.T) {
	var deployID = "center"
	parentKey := ParsePKCS8PEMPrivateKeyFromFile(fmt.Sprintf("../deploy-%s.key", deployID))
	parentCert := ParseX509CertificateFromFile(fmt.Sprintf("../deploy-%s.crt", deployID))
	certDer, privateKey, err := GenerateTLSCertificate(pkix.Name{
		Organization:       []string{deployID},
		OrganizationalUnit: []string{"staff"},
		CommonName:         "Staff CA",
	}, true, parentCert, parentKey, defaultDnsNames)
	if err != nil {
		panic(err)
	}
	var pemEncodedPrivateKey = MarshalRSAPrivateKeyUsingPKCS8PEM(privateKey)
	os.WriteFile(fmt.Sprintf("../%s-staff-ca.key", deployID), pemEncodedPrivateKey, os.ModePerm)
	var pemEncodedCert = MarshalCertificateDerUsingPEM(certDer)
	os.WriteFile(fmt.Sprintf("../%s-staff-ca.crt", deployID), pemEncodedCert, os.ModePerm)
}

func TestGenerateAnonymousCertificate(t *testing.T) {
	var deployID = "center"
	parentKey := ParsePKCS8PEMPrivateKeyFromFile(fmt.Sprintf("../%s-staff-ca.key", deployID))
	parentCert := ParseX509CertificateFromFile(fmt.Sprintf("../%s-staff-ca.crt", deployID))
	certDer, privateKey, err := GenerateTLSCertificate(pkix.Name{
		Organization:       []string{deployID},
		OrganizationalUnit: []string{"staff"},
		CommonName:         "anonymous",
		ExtraNames: []pkix.AttributeTypeAndValue{
			{Type: CertificateTypeOID, Value: "user"},
			{Type: UserTypeOID, Value: "anonymous"},
			{Type: UserIDOID, Value: "anonymous"},
		},
	}, false, parentCert, parentKey, defaultDnsNames)
	if err != nil {
		panic(err)
	}
	var pemEncodedPrivateKey = MarshalRSAPrivateKeyUsingPKCS8PEM(privateKey)
	os.WriteFile(fmt.Sprintf("../%s-anonymous.key", deployID), pemEncodedPrivateKey, os.ModePerm)
	var pemEncodedCert = MarshalCertificateDerUsingPEM(certDer)
	os.WriteFile(fmt.Sprintf("../%s-anonymous.crt", deployID), pemEncodedCert, os.ModePerm)
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
	var pemEncodedPrivateKey = MarshalRSAPrivateKeyUsingPKCS8PEM(privateKey)
	os.WriteFile("../rpc-ca.key", pemEncodedPrivateKey, os.ModePerm)
	var pemEncodedCert = MarshalCertificateDerUsingPEM(certDer)
	os.WriteFile("../rpc-ca.crt", pemEncodedCert, os.ModePerm)
}

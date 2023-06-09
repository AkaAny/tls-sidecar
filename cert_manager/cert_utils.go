package cert_manager

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func ParseX509CertificateFromFile(fileName string) *x509.Certificate {
	var der = ReadSingleDerFromPEMFile(fileName)
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		panic(err)
	}
	//fmt.Println(cert)
	return cert
}

func ParseX509CertificateFromData(data []byte) *x509.Certificate {
	var certDer = ReadSingleDerFromPEMData(data)
	cert, err := x509.ParseCertificate(certDer)
	if err != nil {
		panic(err)
	}
	return cert
}

func ReadSingleDerFromPEMFile(fileName string) []byte {
	fileData, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return ReadSingleDerFromPEMData(fileData)
}

func ReadSingleDerFromPEMData(fileData []byte) []byte {
	pemBlock, _ := pem.Decode(fileData)
	return pemBlock.Bytes
}

func ParsePKCS8PEMPrivateKeyFromFile(fileName string) *rsa.PrivateKey {
	var der = ReadSingleDerFromPEMFile(fileName)
	privateKeyObj, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		panic(err)
	}
	return privateKeyObj.(*rsa.PrivateKey)
}

func ParsePKCS8PEMPrivateKeyFromData(fileData []byte) *rsa.PrivateKey {
	var der = ReadSingleDerFromPEMData(fileData)
	privateKeyObj, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		panic(err)
	}
	return privateKeyObj.(*rsa.PrivateKey)
}

func ParsePEMPrivateKeyFromFile(fileName string) *rsa.PrivateKey {
	var der = ReadSingleDerFromPEMFile(fileName)
	privateKey, err := x509.ParsePKCS1PrivateKey(der)
	if err != nil {
		panic(err)
	}
	return privateKey
}

func MarshalRSAPrivateKeyUsingPKCS8PEM(privateKey *rsa.PrivateKey) (pemData []byte) {
	der, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		panic(err)
	}
	keyPem := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: der,
	}
	return pem.EncodeToMemory(keyPem)
}

func MarshalCertificateDerUsingPEM(der []byte) (pemData []byte) {
	certPem := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: der,
	}
	return pem.EncodeToMemory(certPem)
}

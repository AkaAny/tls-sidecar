package tls_sidecar

import (
	"fmt"
	"testing"
	"tls-sidecar/cert_manager"
)

func TestParsePEMCertificate(t *testing.T) {
	var rpcCert = cert_manager.ParseX509CertificateFromFile("rpc.crt")
	fmt.Println(rpcCert)
}

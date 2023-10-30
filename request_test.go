package tls_sidecar

import (
	tls2 "crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/AkaAny/tls-sidecar/cert_manager"
	"io"
	"net/http"
	"testing"
)

func TestSendRequest(t *testing.T) {
	var deployHDUCert = cert_manager.ParseX509CertificateFromFile("deploy-hdu.crt")
	var rpcCACert = cert_manager.ParseX509CertificateFromFile("rpc-ca.crt")
	//var appCACert = cert_manager.ParseX509CertificateFromFile("app-ca.crt")
	//var rootCACert = cert_manager.ParseX509CertificateFromFile("jl-root-ca.crt")
	err := deployHDUCert.CheckSignatureFrom(rpcCACert)
	if err != nil {
		panic(err)
	}
	tlsCert, err := tls2.LoadX509KeyPair("rpc-service-company.crt", "rpc-service-company.key")
	if err != nil {
		panic(err)
	}
	//tlsCert.Certificate = append(tlsCert.Certificate,
	//	appCACert.Raw,
	//	rootCACert.Raw)
	//tlsCert.Leaf = rpcCert
	var caPool = x509.NewCertPool()
	//caPool.AddCert(rootCACert)
	caPool.AddCert(deployHDUCert)
	//caPool.AddCert(appCACert)
	var customTransport = &http.Transport{
		TLSClientConfig: &tls2.Config{
			Certificates: []tls2.Certificate{
				tlsCert,
			},
			RootCAs: caPool,

			VerifyPeerCertificate: nil,
			VerifyConnection:      nil,
		},
	}
	fmt.Println(customTransport)
	var httpClient = http.Client{Transport: customTransport}
	req, err := http.NewRequest("GET", "https://localhost:30443", nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("X-Target-Deploy-ID", "hdu")
	req.Header.Set("X-Target-Service-ID", "company")
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	responseBodyRawData, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
	fmt.Println(string(responseBodyRawData))
}

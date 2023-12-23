package tls_sidecar

import (
	"bufio"
	tls2 "crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/AkaAny/tls-sidecar/trust_center"
	"github.com/guonaihong/gout"
	"github.com/pkg/errors"
	"net/http"
)

func DoTransfer(tlsConn *tls2.Conn, backendServiceHost string) {
	var bufferedTLSReader = bufio.NewReader(tlsConn)
	r, err := http.ReadRequest(bufferedTLSReader)
	if err != nil {
		panic(err)
	}
	var tlsConnState = tlsConn.ConnectionState()
	r.TLS = &tlsConnState
	fmt.Println("http request:", r)
	fmt.Println("request tls:", r.TLS)
	var peerCert = r.TLS.PeerCertificates[0]
	var peerCertDERB64 = base64.StdEncoding.EncodeToString(peerCert.Raw)
	r.Header.Set("SSL-Client-Cert", peerCertDERB64)
	//var responseWriter = NewResponseWriter(1, 1)
	var rpcIdentityHandler = new(trust_center.RPCIdentityHandler)
	identityHttpHeader, err := rpcIdentityHandler.HandleIdentity(r, peerCert)
	if err != nil {
		panic(err)
	}
	for headerKey, values := range identityHttpHeader {
		r.Header[headerKey] = values
	}
	//var (
	//	method = r.Header.Get("X-Target-Method")
	//	path   = r.Header.Get("X-Target-Path")
	//	query  = r.Header.Get("X-Target-Query")
	//)
	//var realUrlStr = fmt.Sprintf("http://%s%s?%s", h.BackendServiceHost, path, query)
	//fmt.Println("to service real url:", realUrlStr)

	//realUrl, err := url.Parse(realUrlStr)
	//if err != nil {
	//	responseWriter.Header().Set("X-Error", err.Error())
	//	return
	//}
	r.URL.Scheme = "http"
	r.URL.Host = backendServiceHost
	r.RequestURI = "" //can't be set in client requests
	resp, err := gout.New().SetRequest(r).Response()
	if err != nil {
		panic(errors.Wrap(err, "request backend service"))
	}
	resp.Header.Set("X-Request-Url", r.URL.String())
	err = resp.Write(tlsConn)
	//var httpResponse = responseWriter.Response()
	//err = httpResponse.Write(h.tlsConn)
	if err != nil {
		panic(err)
	}
}

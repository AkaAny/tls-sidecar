package tls_sidecar

import (
	"bufio"
	"context"
	"crypto/rsa"
	tls2 "crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/AkaAny/tls-sidecar/trust_center"
	"github.com/guonaihong/gout"
	"github.com/samber/lo"
	"net/http"
	"nhooyr.io/websocket"
)

type WSHandler struct {
	ServiceKey         *rsa.PrivateKey
	ServiceCert        *x509.Certificate
	TrustDeployCerts   []*x509.Certificate
	BackendServiceHost string
}

func (h *WSHandler) Attach(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		panic(err)
	}
	var netConn = websocket.NetConn(context.Background(), c, websocket.MessageBinary)

	var tlsCert = NewTLSCertificate(h.ServiceKey, h.ServiceCert)

	var caPool = x509.NewCertPool()
	lo.ForEach(h.TrustDeployCerts, func(item *x509.Certificate, index int) {
		caPool.AddCert(item)
	})
	//serviceCertInfo, err := trust_center.NewServiceCertificate(h.ServiceCert)
	//if err != nil {
	//	panic(errors.Wrap(err, "invalid service certificate"))
	//}
	var tlsConfig = &tls2.Config{
		Certificates:          []tls2.Certificate{tlsCert},
		GetCertificate:        nil,
		GetClientCertificate:  nil,
		GetConfigForClient:    nil,
		VerifyPeerCertificate: nil,
		VerifyConnection:      nil,
		//RootCAs:               caPool,
		ServerName:         "",
		ClientAuth:         tls2.RequireAndVerifyClientCert, //tls2.RequireAndVerifyClientCert
		ClientCAs:          caPool,
		InsecureSkipVerify: false,
	}
	fmt.Println(tlsConfig)
	var tlsConn = tls2.Server(netConn, tlsConfig)
	var bufferedTLSReader = bufio.NewReader(tlsConn)
	h.handle(tlsConn, bufferedTLSReader) //handle one at a time
}

func (h *WSHandler) handle(tlsConn *tls2.Conn, bufferedTLSReader *bufio.Reader) {
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
	r.URL.Host = h.BackendServiceHost
	r.RequestURI = "" //can't be set in client requests
	resp, err := gout.New().SetRequest(r).Response()
	if err != nil {
		panic(err)
	}
	resp.Header.Set("X-Request-Url", r.URL.String())
	err = resp.Write(tlsConn)
	//var httpResponse = responseWriter.Response()
	//err = httpResponse.Write(h.tlsConn)
	if err != nil {
		panic(err)
	}
	err = tlsConn.Close()
	if err != nil {
		panic(err)
	}
}

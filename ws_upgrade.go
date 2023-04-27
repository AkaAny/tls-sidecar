package tls_sidecar

import (
	"bufio"
	"context"
	"crypto/rsa"
	tls2 "crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/samber/lo"
	"net/http"
	"nhooyr.io/websocket"
	"tls-sidecar/trust_center"
)

type WSHandler struct {
	ServiceKey        *rsa.PrivateKey
	ServiceCert       *x509.Certificate
	TrustDeployCerts  []*x509.Certificate
	tlsConn           *tls2.Conn
	bufferedTLSReader *bufio.Reader
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
		RootCAs:               caPool,
		ServerName:            "",
		ClientAuth:            tls2.RequireAndVerifyClientCert, //tls2.RequireAndVerifyClientCert
		ClientCAs:             caPool,
		InsecureSkipVerify:    false,
	}
	fmt.Println(tlsConfig)
	var tlsConn = tls2.Server(netConn, tlsConfig)
	h.tlsConn = tlsConn
	var bufferedTLSReader = bufio.NewReader(tlsConn)
	h.bufferedTLSReader = bufferedTLSReader
	h.handle() //handle one at a time
}

func (h *WSHandler) handle() {
	r, err := http.ReadRequest(h.bufferedTLSReader)
	if err != nil {
		panic(err)
	}
	var tlsConnState = h.tlsConn.ConnectionState()
	r.TLS = &tlsConnState
	fmt.Println("http request:", r)
	fmt.Println("request tls:", r.TLS)
	var peerCert = r.TLS.PeerCertificates[0]
	var responseWriter = NewResponseWriter(1, 1)
	serviceCertInfo, err := trust_center.NewServiceCertificate(peerCert)
	if err != nil {
		responseWriter.Header().Set("X-Error", err.Error())
	} else {
		responseWriter.Header().Set("X-From-Service", serviceCertInfo.ServiceID)
	}
	responseWriter.Header().Set("X-Request-Url", r.URL.String())
	responseWriter.Write([]byte("this is body"))
	var httpResponse = responseWriter.Response()
	err = httpResponse.Write(h.tlsConn)
	if err != nil {
		panic(err)
	}
	err = h.tlsConn.Close()
	if err != nil {
		panic(err)
	}
}

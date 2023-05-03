package tls_sidecar

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rsa"
	tls2 "crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"nhooyr.io/websocket"
	"strings"
)

type WSClientParam struct {
	TargetWSURL string
	SelfKey     *rsa.PrivateKey
	SelfCert    *x509.Certificate
	DeployCert  *x509.Certificate
}

func NewWSClient(param WSClientParam, request *http.Request) (*http.Response, error) {
	const defaultTargetWSUrl = "ws://localhost:9090/tlsRequest"
	if param.TargetWSURL == "" {
		param.TargetWSURL = defaultTargetWSUrl
	}
	c, _, err := websocket.Dial(context.Background(), param.TargetWSURL, nil)
	if err != nil {
		panic(err)
	}
	//defer c.Close(websocket.StatusInternalError, "the sky is falling")

	var netConn = websocket.NetConn(context.Background(), c, websocket.MessageBinary)

	var tlsCert = NewTLSCertificate(param.SelfKey, param.SelfCert)

	var caPool = x509.NewCertPool()
	//caPool.AddCert(rootCACert)
	caPool.AddCert(param.DeployCert)

	var tlsConfig = &tls2.Config{
		Certificates: []tls2.Certificate{
			tlsCert,
		},
		RootCAs: caPool,

		VerifyPeerCertificate: nil,
		VerifyConnection:      nil,

		ServerName:         "",
		InsecureSkipVerify: true,
	}
	var tlsConn = tls2.Client(netConn, tlsConfig)
	if request == nil {
		defaultRequest, err := http.NewRequest("POST", "http://localhost:9090/abc",
			strings.NewReader("this is request body"))
		if err != nil {
			return nil, errors.Wrap(err, "new http request")
		}
		request = defaultRequest
	}
	if err := request.Write(tlsConn); err != nil {
		panic(err)
	}
	var bufioReader = bufio.NewReader(tlsConn)
	response, err := http.ReadResponse(bufioReader, request)
	if err != nil {
		return nil, errors.Wrap(err, "read http response")
	}
	fmt.Println(response)
	//io.ReadAll(response.Body)
	//if err := netConn.Close(); err != nil {
	//	panic(err)
	//}
	return response, nil
}

func DoTLSRequest(clientParam WSClientParam,
	method, url string, httpHeader http.Header, bodyData []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyData))
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}
	req.Header = httpHeader
	resp, err := NewWSClient(clientParam, req)
	if err != nil {
		return nil, errors.Wrap(err, "send ws data")
	}
	return resp, nil
}

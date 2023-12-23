package tls_on_http

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

func NewHttpClient(serverBaseUrl string, tlsConfig *tls.Config, req *http.Request) (resp *http.Response, err error) {
	var clientConn = NewMockClientConn(serverBaseUrl)
	if err := clientConn.Init(); err != nil {
		panic(err)
	}
	var tlsConn = tls.Client(clientConn, tlsConfig)
	if err := req.Write(tlsConn); err != nil {
		panic(errors.Wrap(err, "write http request"))
	}
	fmt.Println("after write http request to server")
	var bufioReader = bufio.NewReader(tlsConn)
	resp, err = http.ReadResponse(bufioReader, req)
	if err != nil {
		return nil, errors.Wrap(err, "read http response")
	}
	respBodyData, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, errors.Wrap(err, "read response body")
	}
	fmt.Println("resp body:", string(respBodyData))
	var bodyReader = bytes.NewReader(respBodyData)
	resp.Body = io.NopCloser(bodyReader)
	tlsConn.Close()
	return resp, nil
}

package ws_handler

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/AkaAny/tls-sidecar"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"nhooyr.io/websocket"
	"time"
)

type WSClientParam struct {
	TargetWSURL      string
	TLSConfigFactory *tls_sidecar.ClientTLSConfigFactory
}

func NewWSClient(serverBaseUrl string, tlsConfig *tls.Config, request *http.Request) (*http.Response, error) {
	var targetWSURL = fmt.Sprintf("%s/tlsRequest", serverBaseUrl)
	c, _, err := websocket.Dial(context.Background(), targetWSURL, nil)
	if err != nil {
		panic(err)
	}
	//defer c.Close(websocket.StatusInternalError, "the sky is falling")
	ctx, cancelFn := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancelFn()
	var netConn = websocket.NetConn(ctx, c, websocket.MessageBinary)

	var tlsConn = tls.Client(netConn, tlsConfig)
	if err := request.Write(tlsConn); err != nil {
		panic(err)
	}
	fmt.Println("handshake complete status:", tlsConn.ConnectionState().HandshakeComplete)
	var bufioReader = bufio.NewReaderSize(tlsConn, 512)
	response, err := http.ReadResponse(bufioReader, request)
	if err != nil {
		return nil, errors.Wrap(err, "read http response")
	}
	fmt.Println(response)
	respBodyData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read response body")
	}
	var bodyReader = bytes.NewReader(respBodyData)
	response.Body = io.NopCloser(bodyReader)
	tlsConn.Close()
	return response, nil
}

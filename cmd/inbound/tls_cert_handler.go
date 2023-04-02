package inbound

import (
	tls2 "crypto/tls"
	"fmt"
	"github.com/go-netty/go-netty"
	"github.com/pkg/errors"
	"net/http"
)

type TLSCertHandler struct {
}

func (*TLSCertHandler) HandleActive(ctx netty.ActiveContext) {
	fmt.Printf("http client active: %s\n", ctx.Channel().RemoteAddr())
	var channelCtx = ctx.Channel().Context()

	fmt.Println(channelCtx)
	ctx.HandleActive()
}

func (*TLSCertHandler) HandleRead(ctx netty.InboundContext, message netty.Message) {
	request, ok := message.(*http.Request)
	if ok {
		fmt.Printf("[%d]%s: %s %s\n", ctx.Channel().ID(), ctx.Channel().RemoteAddr(), request.Method, request.URL.Path)
	} else {
		panic(errors.Errorf("unknown message:%v", message))
	}
	var rawTransport = ctx.Channel().Transport().RawTransport()
	var tlsConn = rawTransport.(*tls2.Conn)
	var connState = tlsConn.ConnectionState()
	//var peerCert = connState.PeerCertificates[0]
	//send request to real server
	request.TLS = &connState
	//ctx.Channel().SetAttachment(request)
	ctx.HandleRead(message)
}

func (*TLSCertHandler) HandleWrite(ctx netty.OutboundContext, message netty.Message) {
	//if responseWriter, ok := message.(http.ResponseWriter); ok {
	//	// set response header.
	//	responseWriter.Header().Add("x-time", time.Now().String())
	//}
	ctx.HandleWrite(message)
}

func (*TLSCertHandler) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	fmt.Printf("http client inactive: %s %v\n", ctx.Channel().RemoteAddr(), ex)
	ctx.HandleInactive(ex)
}

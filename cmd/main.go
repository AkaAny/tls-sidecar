package main

import (
	tls2 "crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty-transport/tls"
	"github.com/go-netty/go-netty/codec/xhttp"
	"github.com/pkg/errors"
	"net/http"
	tls_sidecar "tls-sidecar"
	"tls-sidecar/cert_manager"
)

func main() {
	fmt.Println("sidecar starts working")
	tlsCert, err := tls2.LoadX509KeyPair("rpc-service-company.crt", "rpc-service-company.key")
	if err != nil {
		panic(err)
	}
	var caPool = x509.NewCertPool()
	//var rootCACert = cert_manager.ParseX509CertificateFromFile("jl-root-ca.crt")
	//caPool.AddCert(rootCACert)
	var deployHDUCert = cert_manager.ParseX509CertificateFromFile("deploy-hdu.crt")
	caPool.AddCert(deployHDUCert)
	//var schoolCACert = cert_manager.ParseX509CertificateFromFile("school-ca.crt")
	//caPool.AddCert(schoolCACert)
	//companyA => hdu, check

	// setup bootstrap & startup server.
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println(request.TLS)

		writer.Write([]byte("Hello, go-netty!"))
	})
	var inboundRouterHandler = &tls_sidecar.ServiceRouteHandler{}

	// child pipeline initializer.
	setupCodec := func(channel netty.Channel) {
		channel.Pipeline().
			// decode http request from channel
			AddLast(xhttp.ServerCodec()).
			// print http access log
			AddLast(new(tlsCertHandler)).
			// compatible with http.Handler
			//AddLast(xhttp.Handler(httpMux))
			AddLast(inboundRouterHandler)
	}
	err = netty.NewBootstrap(netty.WithChildInitializer(setupCodec), netty.WithTransport(tls.New())).
		Listen("0.0.0.0:30443", tls.WithOptions(&tls.Options{
			TLS: &tls2.Config{
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
			},
		})).Sync()
	if err != nil {
		panic(err)
	}
}

type tlsCertHandler struct {
}

func (*tlsCertHandler) HandleActive(ctx netty.ActiveContext) {
	fmt.Printf("http client active: %s\n", ctx.Channel().RemoteAddr())
	var channelCtx = ctx.Channel().Context()

	fmt.Println(channelCtx)
	ctx.HandleActive()
}

func (*tlsCertHandler) HandleRead(ctx netty.InboundContext, message netty.Message) {
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

func (*tlsCertHandler) HandleWrite(ctx netty.OutboundContext, message netty.Message) {
	//if responseWriter, ok := message.(http.ResponseWriter); ok {
	//	// set response header.
	//	responseWriter.Header().Add("x-time", time.Now().String())
	//}
	ctx.HandleWrite(message)
}

func (*tlsCertHandler) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	fmt.Printf("http client inactive: %s %v\n", ctx.Channel().RemoteAddr(), ex)
	ctx.HandleInactive(ex)
}

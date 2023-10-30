package inbound

import (
	"crypto/rsa"
	tls2 "crypto/tls"
	"crypto/x509"
	"fmt"
	tls_sidecar "github.com/AkaAny/tls-sidecar"
	"github.com/AkaAny/tls-sidecar/trust_center"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty-transport/tls"
	"github.com/go-netty/go-netty/codec/xhttp"
	"github.com/pkg/errors"
	"github.com/samber/lo"
)

func Main(trustedDeployCerts []*x509.Certificate,
	serviceCert *x509.Certificate, serviceKey *rsa.PrivateKey,
	serviceHost string) {
	var serverPort = 30443
	fmt.Println("sidecar inbound starts working on port:%d", serverPort)
	var tlsCert = tls_sidecar.NewTLSCertificate(serviceKey, serviceCert)
	//tlsCert, err := tls2.LoadX509KeyPair("rpc-service-company.crt", "rpc-service-company.key")
	//if err != nil {
	//	panic(err)
	//}
	var caPool = x509.NewCertPool()
	lo.ForEach(trustedDeployCerts, func(item *x509.Certificate, index int) {
		caPool.AddCert(item)
	})
	serviceCertInfo, err := trust_center.NewServiceCertificate(serviceCert)
	if err != nil {
		panic(errors.Wrap(err, "invalid service certificate"))
	}
	var inboundRouterHandler = tls_sidecar.NewServiceRouteHandler(serviceCertInfo.ServiceID, serviceHost)

	// child pipeline initializer.
	setupCodec := func(channel netty.Channel) {
		channel.Pipeline().
			// decode http request from channel
			AddLast(xhttp.ServerCodec()).
			// print http access log
			AddLast(new(TLSCertHandler)).
			// compatible with http.Handler
			//AddLast(xhttp.Handler(httpMux))
			AddLast(inboundRouterHandler)
	}
	netty.NewBootstrap(netty.WithChildInitializer(setupCodec), netty.WithTransport(tls.New())).
		Listen(fmt.Sprintf("0.0.0.0:%d", serverPort), tls.WithOptions(&tls.Options{
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
		})).Async(func(err error) {
		panic(errors.Wrap(err, "inbound bootstrap err"))
	})
}

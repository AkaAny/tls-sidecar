package outbound

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/codec/xhttp"
	"github.com/go-netty/go-netty/transport/tcp"
	"github.com/pkg/errors"
	tls_sidecar "tls-sidecar"
)

func Main(selfDeployCert *x509.Certificate,
	serviceCert *x509.Certificate, serviceKey *rsa.PrivateKey,
	deployIDHostMap map[string]string) {
	var serverPort = 30080
	fmt.Println("sidecar outbound starts working on port:", serverPort)

	var outboundWrapperHandler = tls_sidecar.NewTLSWrapperHandler(selfDeployCert,
		serviceCert, serviceKey,
		deployIDHostMap)

	// child pipeline initializer.
	setupCodec := func(channel netty.Channel) {
		channel.Pipeline().
			// decode http request from channel
			AddLast(xhttp.ServerCodec()).
			// compatible with http.Handler
			//AddLast(xhttp.Handler(httpMux))
			AddLast(outboundWrapperHandler)
	}
	netty.NewBootstrap(netty.WithChildInitializer(setupCodec), netty.WithTransport(tcp.New())).
		Listen(fmt.Sprintf("tcp://localhost:%d", serverPort)).Async(func(err error) {
		panic(errors.Wrap(err, "outbound bootstrap err"))
	})
}

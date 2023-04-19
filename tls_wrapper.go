package tls_sidecar

import (
	"crypto/rsa"
	tls2 "crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/codec/xhttp"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"net/http"
	"net/http/httputil"
)

type TLSWrapperHandler struct {
	transport             *http.Transport
	deployIDDeployInfoMap map[string]*DeployInfo
}

func (t *TLSWrapperHandler) GetDeployInfo(deployID string) *DeployInfo {
	deployInfo, ok := t.deployIDDeployInfoMap[deployID]
	if !ok {
		return nil
	}
	return deployInfo
}

func NewTLSWrapperHandler(deployCert *x509.Certificate,
	serviceCert *x509.Certificate, serviceKey *rsa.PrivateKey,
	idDeployHostMap map[string]string) *TLSWrapperHandler {
	var tlsCert = NewTLSCertificate(serviceKey, serviceCert)
	//tlsCert.Certificate = append(tlsCert.Certificate,
	//	appCACert.Raw,
	//	rootCACert.Raw)
	//tlsCert.Leaf = rpcCert
	var caPool = x509.NewCertPool()
	//caPool.AddCert(rootCACert)
	caPool.AddCert(deployCert)
	//caPool.AddCert(appCACert)
	var customTransport = &http.Transport{
		TLSClientConfig: &tls2.Config{
			Certificates: []tls2.Certificate{
				tlsCert,
			},
			RootCAs: caPool,

			VerifyPeerCertificate: nil,
			VerifyConnection:      nil,
		},
	}
	fmt.Println(customTransport)
	var idDeployInfoMap = lo.MapValues(idDeployHostMap, func(host string, deployID string) *DeployInfo {
		return &DeployInfo{
			ID:   deployID,
			Host: host,
			//IDServiceInfoMap: nil,
		}
	})
	return &TLSWrapperHandler{
		transport:             customTransport,
		deployIDDeployInfoMap: idDeployInfoMap,
	}
}

func (t *TLSWrapperHandler) AddDeploy(deployInfos ...*DeployInfo) {
	for _, deployInfo := range deployInfos {
		t.deployIDDeployInfoMap[deployInfo.ID] = deployInfo
	}
}

func (t *TLSWrapperHandler) HandleActive(ctx netty.ActiveContext) {
	ctx.HandleActive()
}

func (t *TLSWrapperHandler) HandleRead(ctx netty.InboundContext, message netty.Message) {
	request, ok := message.(*http.Request)
	if !ok {
		ctx.HandleRead(message)
		return
	}
	fmt.Println(request)
	ctx.Channel().SetAttachment(request)
	var responseWriter = xhttp.NewResponseWriter(request.ProtoMajor, request.ProtoMinor)
	ctx.Channel().Write(responseWriter)
}

func (t *TLSWrapperHandler) HandleWrite(ctx netty.OutboundContext, message netty.Message) {
	request, ok := ctx.Channel().Attachment().(*http.Request)
	if !ok {
		ctx.HandleWrite(message)
		return
	}
	responseWriter, ok := message.(http.ResponseWriter)
	if !ok {
		ctx.HandleWrite(message)
		return
	}
	//init tls config
	var proxyError *StatusError = nil
	var reverseProxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "https"
			targetDeployID,
				targetServiceID :=
				req.Header.Get("X-Target-Deploy-ID"),
				req.Header.Get("X-Target-Service-ID")
			fmt.Println("[target] deploy id:", targetDeployID, "service id:", targetServiceID)
			var deployInfo = t.GetDeployInfo(targetDeployID)
			if deployInfo == nil {
				proxyError = &StatusError{
					error:      errors.Errorf("deploy id:%s does not exist", targetDeployID),
					StatusCode: http.StatusNotFound,
					ErrorCode:  ErrCodeTargetDeployIDNotFound,
				}
				return
			}
			//let inbound handler to check the validation of service and api
			req.Host = deployInfo.Host
			req.URL.Host = deployInfo.Host
			var apiPath = req.URL.Path
			req.URL.Path = "/" +
				targetServiceID + //ingress path prefix must be the same as service id
				apiPath
			req.URL.RawPath = ""
		},
		Transport: NewProxyRoundTripper(t.transport,
			func(base http.RoundTripper, req *http.Request) (*http.Response, error) {
				if proxyError == nil {
					return base.RoundTrip(req)
				}
				return proxyError.AsHttpResponse()
			}),
		FlushInterval:  0,
		ErrorLog:       nil,
		BufferPool:     nil,
		ModifyResponse: nil,
		ErrorHandler:   nil,
	}
	reverseProxy.ServeHTTP(responseWriter, request)
	ctx.HandleWrite(responseWriter)
}

func (t *TLSWrapperHandler) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	ctx.HandleInactive(ex)
}

package tls_sidecar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/codec/xhttp"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"
	"tls-sidecar/trust_center"
)

const (
	ServiceIDDoesNotExist         = "service-404"
	IdentityTypeHeaderKey         = "X-Identity-Type"
	IdentityTypeValueUnauthorized = "unauthorized"
	FromDeployIDHeaderKey         = "X-From-Deploy-ID"
	FromServiceIDHeaderKey        = "X-From-Service-ID"
)

type APIInfo struct {
	FullPath string
	Methods  []string

	SupportedIdentityType []string
}

type ServiceInfo struct {
	PathRouterInfoMap map[string]*APIInfo
	Host              string
}

func (s *ServiceInfo) GetAPIInfoByPath(fullPath string) *APIInfo {
	apiInfo, ok := s.PathRouterInfoMap[fullPath]
	if !ok {
		return nil
	}
	return apiInfo
}

type ServiceRouteHandler struct {
	serviceIDInfoTable map[string]*ServiceInfo
}

func NewServiceRouteHandler(serviceIDHostMap map[string]string) *ServiceRouteHandler {
	//TODO: ask real service to report its api
	var serviceIDInfoMap = lo.MapValues(serviceIDHostMap, func(host string, serviceID string) *ServiceInfo {
		return &ServiceInfo{
			//PathRouterInfoMap: nil,
			Host: host,
		}
	})
	return &ServiceRouteHandler{serviceIDInfoTable: serviceIDInfoMap}
}

func (s *ServiceRouteHandler) GetServiceInfoByServiceID(serviceID string) *ServiceInfo {
	serviceInfo, ok := s.serviceIDInfoTable[serviceID]
	if !ok {
		return nil
	}
	return serviceInfo
}

func (s *ServiceRouteHandler) HandleActive(ctx netty.ActiveContext) {
	ctx.HandleActive()
}

func (s *ServiceRouteHandler) HandleRead(ctx netty.InboundContext, message netty.Message) {
	request, ok := message.(*http.Request)
	if !ok {
		ctx.HandleRead(message)
		return
	}
	ctx.Channel().SetAttachment(request)
	var responseWriter = xhttp.NewResponseWriter(request.ProtoMajor, request.ProtoMinor)
	ctx.Channel().Write(responseWriter)
}

type StatusError struct {
	error
	StatusCode int
	ErrorCode  int
}

func (x *StatusError) AsHttpResponse() (*http.Response, error) {
	var bodyBuffer = bytes.NewBuffer(nil)
	if err := json.NewEncoder(bodyBuffer).Encode(map[string]any{
		"message":   x.Error(),
		"errorCode": x.ErrorCode,
	}); err != nil {
		return nil, errors.Wrap(err, "encode proxy error to json")
	}
	var responseHeader = http.Header{}
	responseHeader.Set("Content-Type", "application/json")
	//responseHeader.Set("Content-Length", fmt.Sprintf("%d", bodyBuffer.Len()))
	return &http.Response{
		StatusCode:    x.StatusCode,
		Header:        responseHeader,
		Body:          io.NopCloser(bodyBuffer),
		ContentLength: int64(bodyBuffer.Len()),
	}, nil
}

func (s *ServiceRouteHandler) HandleWrite(ctx netty.OutboundContext, message netty.Message) {
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
	var proxyError *StatusError = nil
	var reverseProxy = httputil.ReverseProxy{
		Director: func(req *http.Request) {
			var fullPath = req.URL.Path
			var apiPaths = strings.Split(fullPath, "/")[1:]
			var apiPath = strings.Join(apiPaths, "/")
			req.URL.Path = apiPath
			if req.URL.RawPath != "" {
				req.URL.RawPath = apiPath
			}
			var peerCert = request.TLS.PeerCertificates[0]
			fromServiceCertInfo, err := trust_center.NewServiceCertificate(peerCert)
			if err != nil {
				proxyError = &StatusError{
					error:      errors.Wrap(err, "parse service rpc certificate from peer cert"),
					StatusCode: http.StatusBadRequest,
					ErrorCode:  ErrCodeInvalidPeerCertificate,
				}
				return
			}
			req.Header.Set("X-From-Deploy-ID", fromServiceCertInfo.DeployID)
			req.Header.Set("X-From-Service-ID", fromServiceCertInfo.ServiceID)
			deployID, serviceID := req.Header.Get("X-Target-Deploy-ID"), req.Header.Get("X-Target-Service-ID")
			if deployID == "" {
				proxyError = &StatusError{
					error:      errors.New("target deploy id header is missing"),
					StatusCode: http.StatusBadRequest,
					ErrorCode:  ErrCodeMissingTargetDeployIDHeader,
				}
				return
			}
			if serviceID == "" {
				proxyError = &StatusError{
					error:      errors.New("target service id header is missing"),
					StatusCode: http.StatusBadRequest,
					ErrorCode:  ErrCodeMissingTargetServiceIDHeader,
				}
				return
			}
			fmt.Println("[target] deploy id:", deployID, "service id:"+serviceID)
			var targetServiceInfo = s.GetServiceInfoByServiceID(serviceID)
			if targetServiceInfo == nil {
				//req.Host=ServiceIDDoesNotExist
				proxyError = &StatusError{
					error:      errors.Errorf("target service id:%s does not exist", serviceID),
					StatusCode: http.StatusNotFound,
					ErrorCode:  ErrCodeTargetServiceIDNotFound,
				}
				return
			}
			req.Host = targetServiceInfo.Host
			req.URL.Host = targetServiceInfo.Host
			var apiInfo = targetServiceInfo.GetAPIInfoByPath(apiPath)
			if apiInfo == nil {
				proxyError = &StatusError{
					error:      errors.Errorf("api with path:%s does not exist", apiPath),
					StatusCode: http.StatusNotFound,
					ErrorCode:  ErrCodeAPIPathNotFound,
				}
				return
			}
			if !lo.Contains(apiInfo.Methods, req.Method) {
				proxyError = &StatusError{
					error:      errors.Errorf("api method:%s does not exist", req.Method),
					StatusCode: http.StatusNotFound,
					ErrorCode:  ErrCodeAPIMethodNotFound,
				}
				return
			}
			var authHeaderValue = request.Header.Get("Authorization")
			var authHeaderParts = strings.Split(authHeaderValue, " ")
			var identityType = authHeaderParts[0]
			if !lo.Contains(apiInfo.SupportedIdentityType, identityType) {
				proxyError = &StatusError{
					error:      errors.Errorf("unsupported identity type:%s", identityType),
					StatusCode: http.StatusForbidden,
					ErrorCode:  ErrCodeUnsupportedIdentityType,
				}
				return
			}
			switch identityType {
			case "rpc":
				break //入站证书是可信任的，直接放通
			case "jwt":
				var jwtTokenStr = authHeaderParts[1]
				jwtToken, err := jwt.ParseWithClaims(jwtTokenStr, &jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
					return fromServiceCertInfo.Certificate.PublicKey, nil
				})
				if err != nil {
					proxyError = &StatusError{
						error:      errors.Wrap(err, "parse jwt token"),
						StatusCode: http.StatusBadRequest,
						ErrorCode:  ErrCodeFailedToParseJWTToken,
					}
					return
				}
				if !jwtToken.Valid {
					proxyError = &StatusError{
						error:      errors.New("invalid jwt token"),
						StatusCode: http.StatusBadRequest,
						ErrorCode:  ErrCodeInvalidJWTToken,
					}
					return
				}
				mapClaims, ok := jwtToken.Claims.(*jwt.MapClaims)
				if !ok {
					proxyError = &StatusError{
						error:      errors.New("invalid jwt claims"),
						StatusCode: http.StatusBadRequest,
						ErrorCode:  0,
					}
				}
				claimsJsonRawData, _ := json.Marshal(mapClaims)
				req.Header.Set("X-Auth-Claims", string(claimsJsonRawData))
			default:
				req.Header.Set(IdentityTypeHeaderKey, IdentityTypeValueUnauthorized)
			}
		},
		Transport: NewProxyRoundTripper(http.DefaultTransport,
			func(base http.RoundTripper, request *http.Request) (*http.Response, error) {
				if proxyError == nil {
					return base.RoundTrip(request)
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
	//FIXME: if ctx.HandleWrite(message) called here, an EOF will be occurred
	ctx.HandleWrite(responseWriter)
}

func (s *ServiceRouteHandler) HandleInactive(ctx netty.InactiveContext, ex netty.Exception) {
	ctx.HandleInactive(ex)
}

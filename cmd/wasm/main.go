//go:build wasm

package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	tls_sidecar "github.com/AkaAny/tls-sidecar"
	"github.com/AkaAny/tls-sidecar/cert_manager"
	tls_on_http "github.com/AkaAny/tls-sidecar/http_handler"
	"github.com/AkaAny/tls-sidecar/wasm"
	"github.com/AkaAny/tls-sidecar/ws_handler"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
	"syscall/js"
)

func copyFromUInt8Array(arrayJSObj js.Value) []byte {
	var byteLength = arrayJSObj.Get("byteLength").Int()
	var bodyData = make([]byte, byteLength)
	js.CopyBytesToGo(bodyData, arrayJSObj)
	return bodyData
}

func parseTLSParamObject(tlsJSObj js.Value) (serverBaseUrl string, clientTLSConfigFactory *tls_sidecar.ClientTLSConfigFactory) {
	var serverBaseUrlJSObj = tlsJSObj.Get("serverBaseUrl")
	serverBaseUrl = serverBaseUrlJSObj.String()
	var selfKeyDataJSObj = tlsJSObj.Get("selfKey")
	var selfKeyData = selfKeyDataJSObj.String()
	var selfCertDataJSObj = tlsJSObj.Get("selfCert")
	var selfCertData = selfCertDataJSObj.String()
	var parentCertDataJSObj = tlsJSObj.Get("parentCert")
	var parentCertData = parentCertDataJSObj.String()
	clientTLSConfigFactory = &tls_sidecar.ClientTLSConfigFactory{
		SelfKey:    cert_manager.ParsePKCS8PEMPrivateKeyFromData([]byte(selfKeyData)),
		SelfCert:   cert_manager.ParseX509CertificateFromData([]byte(selfCertData)),
		ParentCert: cert_manager.ParseX509CertificateFromData([]byte(parentCertData)),
	}
	return
}

func parseUnderlyingProtocol(serverBaseUrl string) (underlyingProtocol string) {
	serverBaseUri, err := url.Parse(serverBaseUrl)
	if err != nil {
		panic(errors.Wrap(err, "invalid server base url"))
	}
	return serverBaseUri.Scheme
}

func TLSRequest(this js.Value, args []js.Value) interface{} {
	var urlInRequest = args[0].String()
	fmt.Println("url in request:", urlInRequest)
	var requestInfoJSObject = args[1]
	var method = requestInfoJSObject.Get("method").String()
	fmt.Println("method:", method)
	var headersJSObj = requestInfoJSObject.Get("headers")
	var objJSObj = js.Global().Get("Object")
	var headerKeysArray = objJSObj.Call("keys", headersJSObj)
	var httpHeader = make(http.Header)
	for keyIndex := 0; keyIndex < headerKeysArray.Length(); keyIndex++ {
		var headerKey = headerKeysArray.Index(keyIndex).String()
		var headerValue = headersJSObj.Get(headerKey)
		var headerValueStr = ""
		if headerValue.Type() == js.TypeNumber {
			headerValueStr = fmt.Sprintf("%d", headerValue.Int())
		} else {
			headerValueStr = headerValue.String()
		}
		httpHeader.Set(headerKey, headerValueStr)
	}
	fmt.Println("request http header:", httpHeader)
	var bodyUnit8Array = requestInfoJSObject.Get("body")
	var bodyData = copyFromUInt8Array(bodyUnit8Array)
	fmt.Println("request body data:", len(bodyData), string(bodyData))

	req, err := http.NewRequest(method, urlInRequest, bytes.NewReader(bodyData))
	if err != nil {
		panic(errors.Wrap(err, "new request"))
	}
	req.Header = httpHeader

	var tlsJSObj = requestInfoJSObject.Get("tls")
	serverBaseUrl, clientTLSConfigFactory := parseTLSParamObject(tlsJSObj)
	var underlyingProtocol = parseUnderlyingProtocol(serverBaseUrl)
	var promise = wasm.NewGoroutinePromise(func() (js.Value, error) {
		var tlsConfig = clientTLSConfigFactory.NewClientConfig()
		var (
			resp *http.Response
			err  error
		)
		switch underlyingProtocol {
		case "ws":
			fallthrough
		case "wss":
			resp, err = ws_handler.NewWSClient(serverBaseUrl, tlsConfig, req)
			break
		case "http":
			fallthrough
		case "https":
			serverBaseUrl += "/http"
			resp, err = tls_on_http.NewHttpClient(serverBaseUrl, tlsConfig, req)
			break
		default:
			panic(errors.Errorf("invalid underlying protocol:%s", underlyingProtocol))
		}
		if err != nil {
			return js.Value{}, errors.Wrap(err, "do tls request")
		}
		return wrapHttpResponse(resp), nil
	})
	return promise
}

func wrapHttpResponse(resp *http.Response) js.Value {
	var jsHttpResponse = js.Global().Get("Object").New()
	jsHttpResponse.Set("url", resp.Request.URL.String())
	jsHttpResponse.Set("ok", func() js.Value {
		var isOK = resp.StatusCode >= 200 && resp.StatusCode <= 299
		return js.ValueOf(isOK)
	}())
	jsHttpResponse.Set("headers", wasm.WrapHTTPHeader(resp.Header))
	var jsBodyObj = wasm.NewObject()
	var bodyObj = &wasm.ReadableStream{
		ReadCloser: resp.Body,
		Consumed:   false,
	}
	jsHttpResponse.Set("arrayBuffer", js.FuncOf(func(this js.Value, args []js.Value) any {
		return wasm.NewGoroutinePromise(func() (js.Value, error) {
			rawData, err := wasm.ReadAndClose(resp.Body)
			if err != nil {
				return js.Value{}, err
			}
			var rawDataJSArray = wasm.NewUint8Array(rawData)
			return rawDataJSArray, nil
		})
	}))
	jsHttpResponse.Set("text", js.FuncOf(func(this js.Value, args []js.Value) any {
		return wasm.NewGoroutinePromise(func() (js.Value, error) {
			rawData, err := wasm.ReadAndClose(resp.Body)
			if err != nil {
				return js.Value{}, err
			}
			var dataStr = string(rawData)
			return js.ValueOf(dataStr), nil
		})
	}))
	jsHttpResponse.Set("json", js.FuncOf(func(this js.Value, args []js.Value) any {
		return wasm.NewGoroutinePromise(func() (js.Value, error) {
			rawData, err := wasm.ReadAndClose(resp.Body)
			if err != nil {
				return js.Value{}, err
			}
			var dataMap = make(map[string]interface{})
			if err := json.Unmarshal(rawData, &dataMap); err != nil {
				return js.Value{}, fmt.Errorf("unmarshal json")
			}
			return js.ValueOf(dataMap), nil
		})
	}))
	//wrap body

	jsHttpResponse.Set("body", jsBodyObj)
	jsHttpResponse.Set("bodyUsed", js.ValueOf(bodyObj.Consumed))
	return jsHttpResponse
}

func main() {
	var requestWASMObj = js.Global().Get("Object").New()
	requestWASMObj.Set("tlsRequest", js.FuncOf(TLSRequest))
	js.Global().Set("RequestWASM", requestWASMObj)
	var c = make(chan int)
	<-c
}

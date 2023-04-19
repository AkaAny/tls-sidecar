//go:build wasm

package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"syscall/js"
	tls_sidecar "tls-sidecar"
	"tls-sidecar/cert_manager"
	"tls-sidecar/wasm"
)

//go:embed deploy-hdu.crt
var deployCertData []byte

//go:embed rpc-service-company.crt
var serviceCertData []byte

//go:embed rpc-service-company.key
var serviceKeyData []byte

func copyFromUInt8Array(arrayJSObj js.Value) []byte {
	var byteLength = arrayJSObj.Get("byteLength").Int()
	var bodyData = make([]byte, byteLength)
	js.CopyBytesToGo(bodyData, arrayJSObj)
	return bodyData
}

func TLSRequest(this js.Value, args []js.Value) interface{} {
	var url = args[0].String()
	fmt.Println("url:", url)
	var requestInfoJSObject = args[0]
	var method = requestInfoJSObject.Get("method").String()
	fmt.Println("method:", method)
	var headersJSObj = requestInfoJSObject.Get("headers")
	var objJSObj = js.Global().Get("Object")
	var headerKeysArray = objJSObj.Call("keys", headersJSObj)
	var httpHeader = make(http.Header)
	for keyIndex := 0; keyIndex < headerKeysArray.Length(); keyIndex++ {
		var headerKey = headerKeysArray.Index(keyIndex).String()
		var headerValue = headersJSObj.Get(headerKey).Call("toString").String()
		httpHeader.Set(headerKey, headerValue)
	}
	var bodyUnit8Array = requestInfoJSObject.Get("body")
	var bodyData = copyFromUInt8Array(bodyUnit8Array)
	var tlsJSObj = requestInfoJSObject.Get("tls")
	var wsClientParam = func() tls_sidecar.WSClientParam {
		var selfKeyDataJSObj = tlsJSObj.Get("selfKey")
		var selfKeyData = copyFromUInt8Array(selfKeyDataJSObj)
		var selfCertDataJSObj = tlsJSObj.Get("selfCert")
		var selfCertData = copyFromUInt8Array(selfCertDataJSObj)
		var deployCertDataJSObj = tlsJSObj.Get("deployCert")
		var deployCertData = copyFromUInt8Array(deployCertDataJSObj)
		return tls_sidecar.WSClientParam{
			SelfKey:    cert_manager.ParsePKCS8PEMPrivateKeyFromData(selfKeyData),
			SelfCert:   cert_manager.ParseX509CertificateFromData(selfCertData),
			DeployCert: cert_manager.ParseX509CertificateFromData(deployCertData),
		}
	}()
	var promise = wasm.NewPromise(func() (js.Value, error) {
		resp, err := doTLSRequest(wsClientParam, method, url, httpHeader, bodyData)
		if err != nil {
			return js.Value{}, errors.Wrap(err, "do tls request")
		}
		return wrapHttpResponse(resp), nil
	})
	return promise
}

func wrapHttpResponse(resp *http.Response) js.Value {
	var jsHttpResponse = js.Global().Get("Object").New()
	jsHttpResponse.Set("url", resp.Request.URL)
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
	jsHttpResponse.Set("arrayBuffer", wasm.NewPromise(func() (js.Value, error) {
		rawData, err := wasm.ReadAndClose(resp.Body)
		if err != nil {
			return js.Value{}, err
		}
		var rawDataJSArray = wasm.NewUint8Array(rawData)
		return rawDataJSArray, nil
	}))
	jsHttpResponse.Set("text", wasm.NewPromise(func() (js.Value, error) {
		rawData, err := wasm.ReadAndClose(resp.Body)
		if err != nil {
			return js.Value{}, err
		}
		var dataStr = string(rawData)
		return js.ValueOf(dataStr), nil
	}))
	jsHttpResponse.Set("json", wasm.NewPromise(func() (js.Value, error) {
		rawData, err := wasm.ReadAndClose(resp.Body)
		if err != nil {
			return js.Value{}, err
		}
		var dataMap = make(map[string]interface{})
		if err := json.Unmarshal(rawData, &dataMap); err != nil {
			return js.Value{}, fmt.Errorf("unmarshal json")
		}
		return js.ValueOf(dataMap), nil
	}))
	//wrap body

	jsHttpResponse.Set("body", jsBodyObj)
	jsHttpResponse.Set("bodyUsed", js.ValueOf(bodyObj.Consumed))
	return jsHttpResponse
}

func doTLSRequest(clientParam tls_sidecar.WSClientParam,
	method, url string, httpHeader http.Header, bodyData []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyData))
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}
	req.Header = httpHeader
	resp, err := tls_sidecar.NewWSClient(clientParam, req)
	if err != nil {
		return nil, errors.Wrap(err, "send ws data")
	}
	return resp, nil
}

func main() {
	fmt.Println(deployCertData)
	//tls_sidecar.NewWSClient()
	var deployCert = cert_manager.ParseX509CertificateFromData(deployCertData)
	var serviceCert = cert_manager.ParseX509CertificateFromData(serviceCertData)
	var serviceKey = cert_manager.ParsePKCS8PEMPrivateKeyFromData(serviceKeyData)
	tls_sidecar.NewWSClient(tls_sidecar.WSClientParam{
		SelfKey:    serviceKey,
		SelfCert:   serviceCert,
		DeployCert: deployCert,
	}, nil)
	fmt.Println("response")
	js.Global().Set("tlsRequest", js.FuncOf(TLSRequest))
	js.Global().Set("ready", js.ValueOf(true))
	var c = make(chan int)
	<-c
}

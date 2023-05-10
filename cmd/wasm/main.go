//go:build wasm

package main

import (
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
		var targetWSURLJSObj = tlsJSObj.Get("targetWSURL")
		var targetWSURL = targetWSURLJSObj.String()
		var selfKeyDataJSObj = tlsJSObj.Get("selfKey")
		var selfKeyData = selfKeyDataJSObj.String()
		var selfCertDataJSObj = tlsJSObj.Get("selfCert")
		var selfCertData = selfCertDataJSObj.String()
		var parentCertDataJSObj = tlsJSObj.Get("parentCert")
		var parentCertData = parentCertDataJSObj.String()
		return tls_sidecar.WSClientParam{
			TargetWSURL: targetWSURL,
			SelfKey:     cert_manager.ParsePKCS8PEMPrivateKeyFromData([]byte(selfKeyData)),
			SelfCert:    cert_manager.ParseX509CertificateFromData([]byte(selfCertData)),
			ParentCert:  cert_manager.ParseX509CertificateFromData([]byte(parentCertData)),
		}
	}()
	var promise = wasm.NewGoroutinePromise(func() (js.Value, error) {
		resp, err := tls_sidecar.DoTLSRequest(wsClientParam, method, url, httpHeader, bodyData)
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
	jsHttpResponse.Set("arrayBuffer", wasm.NewGoroutinePromise(func() (js.Value, error) {
		rawData, err := wasm.ReadAndClose(resp.Body)
		if err != nil {
			return js.Value{}, err
		}
		var rawDataJSArray = wasm.NewUint8Array(rawData)
		return rawDataJSArray, nil
	}))
	jsHttpResponse.Set("text", wasm.NewGoroutinePromise(func() (js.Value, error) {
		rawData, err := wasm.ReadAndClose(resp.Body)
		if err != nil {
			return js.Value{}, err
		}
		var dataStr = string(rawData)
		return js.ValueOf(dataStr), nil
	}))
	jsHttpResponse.Set("json", wasm.NewGoroutinePromise(func() (js.Value, error) {
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

func main() {
	var requestWASMObj = js.Global().Get("Object").New()
	requestWASMObj.Set("tlsRequest", js.FuncOf(TLSRequest))
	js.Global().Set("RequestWASM", requestWASMObj)
	var c = make(chan int)
	<-c
}

package wasm

import (
	"github.com/samber/lo"
	"net/http"
	"syscall/js"
)

func WrapHTTPHeader(httpHeader http.Header) js.Value {
	var headerJSObj = NewObject()
	for k, values := range httpHeader {
		var valuesInterfaces = lo.Map(values, func(item string, index int) interface{} {
			return item
		})
		headerJSObj.Set(k, js.ValueOf(valuesInterfaces))
	}
	return headerJSObj
}

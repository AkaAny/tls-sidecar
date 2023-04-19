package wasm

import "syscall/js"

func NewUint8Array(rawData []byte) js.Value {
	var rawDataJSArray = js.Global().Get("Uint8Array").New(len(rawData))
	js.CopyBytesToJS(rawDataJSArray, rawData)
	return rawDataJSArray
}

package wasm

import (
	"io"
	"syscall/js"
)

type ReadableStream struct {
	ReadCloser io.ReadCloser
	Consumed   bool
}

func (r *ReadableStream) ReadJS(this js.Value, args []js.Value) interface{} {
	return NewPromise(func() (js.Value, error) {
		var resultObj = NewObject()
		resultObj.Set("done", js.ValueOf(true))
		rawData, err := ReadAndClose(r.ReadCloser)
		if err != nil {
			return js.Value{}, err
		}
		var rawDataJSArray = NewUint8Array(rawData)
		resultObj.Set("value", rawDataJSArray)
		r.Consumed = true
		return rawDataJSArray, nil
	})
}

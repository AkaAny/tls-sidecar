//go:build wasm

package wasm

import (
	"syscall/js"
)

func NewPromise(actionFunc func() (js.Value, error)) js.Value {
	var promiseFunc = func(this js.Value, args []js.Value) interface{} {
		resolve, reject := args[0], args[1]
		res, err := actionFunc()
		if err != nil {
			reject.Invoke(js.Error{
				Value: js.ValueOf(err.Error()),
			})
			return nil
		}
		resolve.Invoke(res)
		return nil
	}
	var promise = js.Global().Get("Promise").New(js.FuncOf(promiseFunc))
	return promise
}

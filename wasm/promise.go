//go:build wasm

package wasm

import (
	"syscall/js"
)

func NewGoroutinePromise(actionFunc func() (js.Value, error)) js.Value {
	var promiseFunc = func(this js.Value, args []js.Value) interface{} {
		resolve, reject := args[0], args[1]
		go func() {
			res, err := actionFunc()
			if err != nil {
				var errJSObj = js.Global().Get("Error").New(err.Error())
				reject.Invoke(errJSObj)
				return
			}
			resolve.Invoke(res)
		}()
		return nil
	}
	var promise = js.Global().Get("Promise").New(js.FuncOf(promiseFunc))
	return promise
}

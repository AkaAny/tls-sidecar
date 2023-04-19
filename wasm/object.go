package wasm

import "syscall/js"

func NewObject() js.Value {
	return js.Global().Get("Object").New()
}

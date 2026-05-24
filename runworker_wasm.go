//go:build js && wasm

package gutter

import "syscall/js"

// RunWorker is the worker-side counterpart to RunApp. Build your worker
// as its own `package main` and compile it to a separate WASM file:
//
//	GOOS=js GOARCH=wasm go build -o worker.wasm ./worker
//
// Then load it from the main app:
//
//	widgets.Worker{
//	    WASM:    "worker.wasm",
//	    Message: "42",
//	    Builder: func(snap widgets.WorkerSnapshot) gutter.Widget { ... },
//	}
//
// handler is invoked once per incoming message; its return value is
// posted back to the main thread as the worker's reply. RunWorker
// installs the message listener and blocks the goroutine so Go's
// scheduler keeps delivering JS callbacks — same shape as RunApp.
//
// Communication is string-only on purpose: it matches widgets.Worker's
// snapshot API and keeps the boundary trivially serializable. Pass JSON
// (or any other text format) for structured data.
func RunWorker(handler func(msg string) string) {
	self := js.Global()
	cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var msg string
		if len(args) > 0 {
			// The handler may be invoked two ways: by the browser
			// dispatching a MessageEvent (args[0] is the event object
			// with a .data field) or by our own drain loop passing the
			// buffered data directly (args[0] is the payload itself,
			// usually a string). Only call .Get on actual objects —
			// js.Value.Get panics on strings/numbers.
			data := args[0]
			if data.Type() == js.TypeObject {
				if d := data.Get("data"); !d.IsUndefined() && !d.IsNull() {
					data = d
				}
			}
			switch data.Type() {
			case js.TypeString:
				msg = data.String()
			case js.TypeUndefined, js.TypeNull:
				// leave msg empty
			default:
				msg = self.Get("JSON").Call("stringify", data).String()
			}
		}
		reply := handler(msg)
		self.Call("postMessage", reply)
		return nil
	})

	// Web Workers drop incoming messages if no handler is installed by
	// the time they arrive — the spec only queues for MessagePort, not
	// for the worker's implicit port when set via onmessage =. So our
	// widgets.Worker bootstrap registers a pre-handler that buffers
	// messages into __GUTTER_QUEUE. Drain it here, then take over.
	if queue := self.Get("__GUTTER_QUEUE"); !queue.IsUndefined() && !queue.IsNull() {
		n := queue.Length()
		for i := 0; i < n; i++ {
			cb.Invoke(queue.Index(i))
		}
		if pre := self.Get("__GUTTER_PRE_HANDLER"); pre.Truthy() {
			self.Call("removeEventListener", "message", pre)
		}
		self.Delete("__GUTTER_PRE_HANDLER")
		self.Delete("__GUTTER_QUEUE")
	}
	self.Set("onmessage", cb)
	select {}
}

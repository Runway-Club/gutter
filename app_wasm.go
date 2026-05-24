//go:build js && wasm

package gutter

import "syscall/js"

// App is the running application instance. It owns the root Element of the
// reconciliation tree and the DOM container it lives in. Per-element state
// (DOM nodes, event listeners, State objects) is held inside each Element, so
// the App itself stays small.
type App struct {
	container js.Value
	root      Element
	ctx       *BuildContext
}

// RunApp mounts root inside the configured container (defaults to #app) using
// the configured theme (defaults to themes.Default) and blocks the main
// goroutine so Go's scheduler can keep delivering JS callbacks.
//
// Pass gutter.WithTheme(themes.Apple) or gutter.WithSelector("#root") to
// customize either, in any order.
//
// RunApp also doubles as the worker entry point: if widgets.Worker
// booted this WASM in a Web Worker context, the bootstrap will have set
// self.__GUTTER_WORKER_TASK to the task name. RunApp detects that
// before any DOM access and dispatches to the registered handler via
// RunWorker, so a single main() — `gutter.RunApp(MyApp{})` — runs both
// roles. Tasks are registered with gutter.NewWorkerTask at package init.
func RunApp(root Widget, opts ...Option) {
	if name := workerTaskHint(); name != "" {
		dispatchWorker(name)
		return
	}
	cfg := newRunConfig(opts)
	doc := js.Global().Get("document")
	container := doc.Call("querySelector", cfg.selector)
	if container.IsNull() || container.IsUndefined() {
		panic("gutter: container not found for selector " + cfg.selector)
	}
	a := &App{
		container: container,
		ctx:       &BuildContext{Theme: cfg.theme},
	}
	a.root = newElement(root)
	a.root.mount(container, js.Null(), a.ctx)
	select {}
}

// workerTaskHint returns the task name the bootstrap wrote into the
// worker global, or "" if we're running in the document context.
func workerTaskHint() string {
	v := js.Global().Get("__GUTTER_WORKER_TASK")
	if v.IsUndefined() || v.IsNull() {
		return ""
	}
	return v.String()
}

// dispatchWorker hands the worker thread to the registered handler, or
// posts an error back to the main thread if the name is unknown — the
// app must register the same name on both sides, so an unknown name is
// always a programming mistake.
func dispatchWorker(name string) {
	handler := lookupWorkerTask(name)
	if handler == nil {
		self := js.Global()
		cb := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			self.Call("postMessage", "gutter: unregistered worker task "+name)
			return nil
		})
		self.Set("onmessage", cb)
		select {}
	}
	RunWorker(handler)
}

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
	MountInto(cfg.selector, root, opts...)
	select {}
}

// MountInto mounts root into the element matched by selector and returns the
// running App. Unlike RunApp it does NOT block — call it once per island, then
// keep the program alive yourself (a trailing `select {}` in main). This is the
// basis of "islands": several independent Gutter trees embedded in an existing
// HTML page, each owning a placeholder element.
//
// Honors WithTheme and WithHydrate (a placeholder containing server-rendered
// markup is hydrated; an empty one is mounted fresh). The selector argument
// takes precedence over WithSelector.
func MountInto(selector string, root Widget, opts ...Option) *App {
	cfg := newRunConfig(opts)
	doc := js.Global().Get("document")
	container := doc.Call("querySelector", selector)
	if container.IsNull() || container.IsUndefined() {
		panic("gutter: container not found for selector " + selector)
	}
	a := &App{
		container: container,
		ctx:       &BuildContext{Theme: cfg.theme},
	}
	a.root = newElement(root)
	// Hydrate server-rendered DOM when asked and present; otherwise mount fresh.
	if children := container.Get("children"); cfg.hydrate && !children.IsUndefined() && children.Get("length").Int() > 0 {
		a.root.hydrate(children.Index(0), a.ctx)
	} else {
		a.root.mount(container, js.Null(), a.ctx)
	}
	registerApp(a) // track for Inspect()/EnableDevtools
	return a
}

// MountWhenVisible defers MountInto until the target element first scrolls into
// (or near) the viewport, via an IntersectionObserver. Use it for below-the-fold
// islands so their build/mount cost is paid only if the user reaches them. The
// WASM module must already be running; to also defer the WASM *download*, pair
// this with the island loader snippet that lazy-loads app.wasm on first
// visibility (see the islands example).
func MountWhenVisible(selector string, root Widget, opts ...Option) {
	doc := js.Global().Get("document")
	el := doc.Call("querySelector", selector)
	if el.IsNull() || el.IsUndefined() {
		panic("gutter: container not found for selector " + selector)
	}
	var obs js.Value
	var cb js.Func
	cb = js.FuncOf(func(_ js.Value, args []js.Value) any {
		entries := args[0]
		if entries.Get("length").Int() > 0 && entries.Index(0).Get("isIntersecting").Bool() {
			obs.Call("disconnect")
			cb.Release()
			MountInto(selector, root, opts...)
		}
		return nil
	})
	obs = js.Global().Get("IntersectionObserver").New(cb)
	obs.Call("observe", el)
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

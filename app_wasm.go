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
func RunApp(root Widget, opts ...Option) {
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

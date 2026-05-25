//go:build js && wasm

package widgets

import "syscall/js"

// initialPath reads the browser's current path (plus query string) so a fresh
// page load lands on the route the user requested.
func initialPath() string {
	loc := js.Global().Get("window").Get("location")
	p := loc.Get("pathname").String()
	if s := loc.Get("search").String(); s != "" {
		p += s
	}
	return p
}

// installHistoryListener wires popstate so the browser back/forward buttons
// drive the router. The listener lives for the lifetime of the page; we do
// not release the js.Func because Router has no Dispose.
func (r *Router) installHistoryListener() {
	cb := js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		loc := js.Global().Get("window").Get("location")
		p := loc.Get("pathname").String()
		if s := loc.Get("search").String(); s != "" {
			p += s
		}
		r.navigated(p)
		return nil
	})
	js.Global().Get("window").Call("addEventListener", "popstate", cb)
}

func (r *Router) pushHistory(path string) {
	js.Global().Get("window").Get("history").Call("pushState", nil, "", path)
}

func (r *Router) replaceHistory(path string) {
	js.Global().Get("window").Get("history").Call("replaceState", nil, "", path)
}

func (r *Router) popHistory() {
	js.Global().Get("window").Get("history").Call("back")
}

//go:build js && wasm

package widgets

import "syscall/js"

// attachScrollListener registers a passive "scroll" listener on the
// viewport node and invokes onScroll with the current scrollTop and
// clientHeight on every event — plus once synchronously after mount so
// ListBuilder can read the real viewport size before its first paint.
//
// Returns an idempotent cleanup that removes the listener and releases the
// js.Func allocation.
func attachScrollListener(node any, onScroll func(scrollTop, viewportHeight float64)) func() {
	n, ok := node.(js.Value)
	if !ok || onScroll == nil {
		return func() {}
	}
	released := false
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, _ []js.Value) any {
		onScroll(n.Get("scrollTop").Float(), n.Get("clientHeight").Float())
		return nil
	})
	// passive: true tells Chrome we won't preventDefault, so it can pipe
	// scroll-handling through the compositor without waiting on Go.
	opts := js.Global().Get("Object").New()
	opts.Set("passive", true)
	n.Call("addEventListener", "scroll", cb, opts)

	// Fire once synchronously so the State picks up the real clientHeight
	// before the first repaint. The "scroll" event itself fires only when
	// scrollTop actually changes, so without this we'd be stuck with the
	// first-render fallback viewport size.
	onScroll(n.Get("scrollTop").Float(), n.Get("clientHeight").Float())

	return func() {
		if released {
			return
		}
		released = true
		n.Call("removeEventListener", "scroll", cb)
		cb.Release()
	}
}

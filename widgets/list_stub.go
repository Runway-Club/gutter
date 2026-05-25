//go:build !js || !wasm

package widgets

// attachScrollListener is a no-op on host builds. ListBuilder still
// renders its first-frame fallback layout, but virtualization needs the
// browser's scroll events which only exist under GOOS=js GOARCH=wasm.
func attachScrollListener(node any, horizontal bool, onScroll func(offset, viewport float64)) func() {
	return func() {}
}

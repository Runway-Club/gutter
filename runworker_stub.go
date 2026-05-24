//go:build !js || !wasm

package gutter

// RunWorker is the worker entry point. The non-WASM stub panics so
// non-wasm builds fail fast at runtime if executed; the binary is only
// meaningful when built with GOOS=js GOARCH=wasm.
func RunWorker(handler func(msg string) string) {
	panic("gutter: RunWorker is only available when built with GOOS=js GOARCH=wasm")
}

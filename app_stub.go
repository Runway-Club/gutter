//go:build !js || !wasm

package gutter

// This stub exists so that user code (which calls RunApp from a main package)
// compiles on the host platform too — handy for editor analysis, `go vet`, and
// any non-WASM tooling. Actually invoking RunApp outside js/wasm is a
// programmer error; the binary is meant to be built with GOOS=js GOARCH=wasm.

// RunApp is the WASM entry point. The non-WASM stub panics so non-wasm builds
// fail fast at runtime if they are ever executed.
func RunApp(root Widget, opts ...Option) {
	panic("gutter: RunApp is only available when built with GOOS=js GOARCH=wasm")
}

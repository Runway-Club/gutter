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

// App is the host-side placeholder for the running application instance. The
// real type (with DOM handles) lives in app_wasm.go.
type App struct{}

// MountInto is the host stub for the islands multi-root mount entry point.
func MountInto(selector string, root Widget, opts ...Option) *App {
	panic("gutter: MountInto is only available when built with GOOS=js GOARCH=wasm")
}

// MountWhenVisible is the host stub for the viewport-lazy island mount.
func MountWhenVisible(selector string, root Widget, opts ...Option) {
	panic("gutter: MountWhenVisible is only available when built with GOOS=js GOARCH=wasm")
}

// Transition just runs fn on host builds — there is no rebuild scheduler during
// SSR, so the priority distinction is meaningful only under GOOS=js GOARCH=wasm.
func Transition(fn func()) { fn() }

//go:build !js || !wasm

package gutter

// SetTitle is a no-op outside the WASM target — there's no document on the
// host platform. The function exists so that widget packages don't need
// build tags.
func SetTitle(title string) {}

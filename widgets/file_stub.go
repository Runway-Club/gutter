//go:build !js || !wasm

package widgets

// attachFileChangeListener is a no-op on non-WASM builds. The File widget
// renders the same DOM-less label but never invokes OnSelect because file
// reading depends on the browser's FileReader API.
func attachFileChangeListener(node any, onSelect func([]FilePick)) func() {
	return func() {}
}

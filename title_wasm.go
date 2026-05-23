//go:build js && wasm

package gutter

import "syscall/js"

// SetTitle updates document.title. Safe to call repeatedly; Scaffold calls
// this whenever it builds with a non-empty Title.
func SetTitle(title string) {
	js.Global().Get("document").Set("title", title)
}

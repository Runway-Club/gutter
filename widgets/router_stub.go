//go:build !js || !wasm

package widgets

// Host-build stubs so `go vet ./...` and editor analysis work outside WASM.
// None of these are functional — Router is only meaningful in the browser.

func initialPath() string                    { return "/" }
func (r *Router) installHistoryListener()    {}
func (r *Router) pushHistory(path string)    {}
func (r *Router) replaceHistory(path string) {}
func (r *Router) popHistory()                {}

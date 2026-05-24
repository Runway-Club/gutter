//go:build !js || !wasm

package widgets

// workerHandle is a no-op on the host platform. The Worker widget only
// does meaningful work under GOOS=js GOARCH=wasm; on host builds the
// type and helpers exist so user code that constructs a Worker still
// compiles for editor tooling and `go vet`.
type workerHandle struct{}

// resolveURL is the host stub for the WASM-only URL resolver used by
// the bootstrap generator. Host builds never spawn workers, so we just
// return the reference unchanged.
func resolveURL(ref string) string { return ref }

func newWorkerHandle(w Worker, onMsg, onErr func(string)) *workerHandle {
	return &workerHandle{}
}

func (h *workerHandle) post(msg string)  {}
func (h *workerHandle) terminate()        {}

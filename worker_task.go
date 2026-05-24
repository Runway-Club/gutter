package gutter

import "sync"

// WorkerTask is a token returned by NewWorkerTask. Pass it to
// widgets.Worker.Task and the widget will spawn a Web Worker that
// dispatches messages to the registered handler. The token is
// cross-platform — it's a plain struct, no JS handles inside.
type WorkerTask struct {
	// Name is the unique dispatch key. The Worker bootstrap writes this
	// to self.__GUTTER_WORKER_TASK before loading the WASM so the
	// worker-side RunApp knows which handler to invoke.
	Name string
	// URL is the worker WASM to load. Defaults to "app.wasm" — the
	// single-binary pattern, where the same WASM doubles as the app and
	// every worker. Override only if you genuinely want a separate
	// binary.
	URL string
}

var (
	workerRegistryMu sync.RWMutex
	workerRegistry   = map[string]func(string) string{}
)

// NewWorkerTask registers handler under name and returns a token to wire
// into widgets.Worker.Task. The handler runs inside a Web Worker the
// first time the widget mounts; subsequent posts re-use the same Worker.
//
// The handler body lives in the main app binary, so it can call any
// helper defined in your project. It must not touch the DOM (workers
// have no document) and must not depend on init-time side effects that
// only happen in the UI context — workers run main() too, but RunApp
// short-circuits before it mounts.
//
// Name must be unique within the app and stable across builds.
func NewWorkerTask(name string, handler func(msg string) string) WorkerTask {
	if name == "" {
		panic("gutter.NewWorkerTask: name is required")
	}
	if handler == nil {
		panic("gutter.NewWorkerTask: handler is required")
	}
	workerRegistryMu.Lock()
	if _, dup := workerRegistry[name]; dup {
		workerRegistryMu.Unlock()
		panic("gutter.NewWorkerTask: duplicate task name " + name)
	}
	workerRegistry[name] = handler
	workerRegistryMu.Unlock()
	return WorkerTask{Name: name, URL: "app.wasm"}
}

// lookupWorkerTask is used by the WASM-side RunApp dispatcher.
func lookupWorkerTask(name string) func(string) string {
	workerRegistryMu.RLock()
	defer workerRegistryMu.RUnlock()
	return workerRegistry[name]
}

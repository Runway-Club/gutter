package widgets

import (
	"fmt"

	"github.com/Runway-Club/gutter"
)

// Worker runs a piece of heavy work in a Web Worker so the main UI thread
// keeps responding. The widget owns one Worker instance for its lifetime
// and exposes the latest result through a builder, the same way
// Flutter's FutureBuilder exposes a Future.
//
// Source modes, in priority order:
//
//   - Task: a gutter.WorkerTask returned by gutter.NewWorkerTask. The
//     handler body lives in the main app binary; the widget spawns a
//     Worker that reloads the same WASM with self.__GUTTER_WORKER_TASK
//     set, and gutter.RunApp dispatches to the registered handler
//     instead of mounting the DOM. This is the recommended mode — the
//     heavy task is plain Go in your app, no separate file, no JS.
//
//   - WASM: path to a separate Go program built with RunWorker.
//     Useful when you want a small purpose-built worker binary instead
//     of reloading the whole app.
//
//   - ScriptURL: classic JS worker file served by the app.
//
//   - Inline: raw JavaScript source, wrapped in a Blob URL.
//
// Messages between the Go side and the worker are strings; if you need
// to pass structured data, pick a serialization (JSON is conventional)
// and parse on both ends.
//
// Builder is invoked on every rebuild with the current snapshot. The
// initial snapshot is Pending=true; once a reply arrives, Message holds
// the payload and Pending becomes false. The Post function inside the
// snapshot lets the UI send follow-up messages — call it from a button
// handler or any other event.
//
// The Worker is created in InitState and terminated in Dispose. Changing
// the source or Message after mount has no effect — the existing worker
// stays attached to its original script. To restart with a new script,
// wrap the Worker in a Keyed widget and change the key.
type Worker struct {
	Task        gutter.WorkerTask
	WASM        string
	WASMExecURL string
	ScriptURL   string
	Inline      string
	Message     string
	Builder     func(WorkerSnapshot) gutter.Widget
}

// WorkerSnapshot is the input to Worker.Builder. Pending is true between
// posting a message and receiving the next reply. Message holds the most
// recent reply (string). Error holds the most recent error message, if
// any. Post sends a fresh message to the worker and flips Pending back to
// true; it is safe to call from event handlers.
type WorkerSnapshot struct {
	Pending bool
	Message string
	Error   string
	Post    func(string)
}

func (w Worker) CreateState() gutter.State {
	return &workerState{widget: w}
}

type workerState struct {
	gutter.StateObject
	widget Worker
	handle *workerHandle
	snap   WorkerSnapshot
}

func (s *workerState) InitState() {
	s.snap = WorkerSnapshot{Pending: s.widget.Message != "", Post: s.post}
	s.handle = newWorkerHandle(resolveSource(s.widget), s.onMessage, s.onError)
	if s.widget.Message != "" {
		s.handle.post(s.widget.Message)
	}
}

// resolveSource picks one source mode and returns a Worker with
// ScriptURL or Inline filled in for the handle layer. Precedence:
// Task > WASM > ScriptURL > Inline. The Task and WASM cases expand to a
// tiny JS bootstrap that loads wasm_exec.js (via importScripts) and
// instantiates the Go program. For Task, the bootstrap also writes
// __GUTTER_WORKER_TASK before loading so gutter.RunApp dispatches to
// the registered handler instead of mounting the DOM.
func resolveSource(w Worker) Worker {
	exec := w.WASMExecURL
	if exec == "" {
		exec = "wasm_exec.js"
	}
	// Pre-resolve to absolute URLs. The bootstrap runs in a blob:
	// origin where relative URLs resolve against the blob, not the
	// page — passing the raw "wasm_exec.js" makes importScripts throw
	// "URL is invalid".
	exec = resolveURL(exec)
	switch {
	case w.Task.Name != "":
		wasm := w.Task.URL
		if wasm == "" {
			wasm = "app.wasm"
		}
		wasm = resolveURL(wasm)
		w.Inline = fmt.Sprintf(
			"self.__GUTTER_WORKER_TASK=%q;%simportScripts(%q);const go=new Go();"+
				"fetch(%q).then(r=>r.arrayBuffer())"+
				".then(b=>WebAssembly.instantiate(b,go.importObject))"+
				".then(r=>go.run(r.instance));",
			w.Task.Name, preHandlerJS, exec, wasm,
		)
		w.ScriptURL = ""
	case w.WASM != "":
		wasm := resolveURL(w.WASM)
		w.Inline = fmt.Sprintf(
			"%simportScripts(%q);const go=new Go();fetch(%q).then(r=>r.arrayBuffer())"+
				".then(b=>WebAssembly.instantiate(b,go.importObject))"+
				".then(r=>go.run(r.instance));",
			preHandlerJS, exec, wasm,
		)
		w.ScriptURL = ""
	}
	return w
}

// preHandlerJS buffers messages that arrive before Go finishes booting.
// Web Workers do not queue messages for an onmessage handler installed
// after the message arrives — they're dropped. The Go side
// (gutter.RunWorker) drains __GUTTER_QUEUE the moment it's ready and
// removes this pre-handler before installing its own.
const preHandlerJS = `self.__GUTTER_QUEUE=[];self.__GUTTER_PRE_HANDLER=function(e){self.__GUTTER_QUEUE.push(e.data);};self.addEventListener("message",self.__GUTTER_PRE_HANDLER);`

func (s *workerState) Dispose() {
	if s.handle != nil {
		s.handle.terminate()
		s.handle = nil
	}
}

func (s *workerState) Build(ctx *gutter.BuildContext) gutter.Widget {
	if s.widget.Builder == nil {
		return Text{}
	}
	return s.widget.Builder(s.snap)
}

func (s *workerState) post(msg string) {
	if s.handle == nil {
		return
	}
	s.SetState(func() {
		s.snap.Pending = true
		s.snap.Error = ""
	})
	s.handle.post(msg)
}

func (s *workerState) onMessage(msg string) {
	s.SetState(func() {
		s.snap.Pending = false
		s.snap.Message = msg
		s.snap.Error = ""
	})
}

func (s *workerState) onError(err string) {
	s.SetState(func() {
		s.snap.Pending = false
		s.snap.Error = err
	})
}

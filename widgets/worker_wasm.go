//go:build js && wasm

package widgets

import "syscall/js"

// resolveURL resolves ref against the document's location.href and
// returns the absolute URL. The worker bootstrap runs inside a blob:
// origin and cannot resolve relative URLs against the page — by the
// time importScripts("wasm_exec.js") runs, the worker's own base URL
// is the blob, not the page, so the browser rejects the relative form
// with "URL is invalid". We pre-resolve in the main thread (where
// location.href is the real page URL) and embed the absolute string.
func resolveURL(ref string) string {
	location := js.Global().Get("location")
	if location.IsUndefined() || location.IsNull() {
		return ref
	}
	href := location.Get("href")
	if href.IsUndefined() || href.IsNull() {
		return ref
	}
	u := js.Global().Get("URL").New(ref, href)
	return u.Get("href").String()
}

// workerHandle wraps the live JS Worker object plus the listener
// closures we have to keep alive until terminate. Inline scripts get a
// Blob URL that is revoked on terminate so we don't leak it.
type workerHandle struct {
	worker    js.Value
	blobURL   string
	onMessage js.Func
	onError   js.Func
	dead      bool
}

func newWorkerHandle(w Worker, onMsg, onErr func(string)) *workerHandle {
	h := &workerHandle{}
	var scriptURL string
	switch {
	case w.ScriptURL != "":
		scriptURL = w.ScriptURL
	case w.Inline != "":
		blob := js.Global().Get("Blob").New(
			[]any{w.Inline},
			map[string]any{"type": "application/javascript"},
		)
		h.blobURL = js.Global().Get("URL").Call("createObjectURL", blob).String()
		scriptURL = h.blobURL
	default:
		// No script provided. Surface the misconfiguration through the
		// snapshot so the UI shows the problem instead of silently doing
		// nothing.
		onErr("widgets.Worker: ScriptURL or Inline is required")
		return h
	}

	h.worker = js.Global().Get("Worker").New(scriptURL)
	h.onMessage = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var payload string
		if len(args) > 0 {
			data := args[0].Get("data")
			if !data.IsUndefined() && !data.IsNull() {
				if data.Type() == js.TypeString {
					payload = data.String()
				} else {
					// Non-string payloads are JSON-encoded so the Go side
					// always sees a string. Callers that want a typed
					// payload can json.Unmarshal it.
					payload = js.Global().Get("JSON").Call("stringify", data).String()
				}
			}
		}
		onMsg(payload)
		return nil
	})
	h.onError = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var msg string
		if len(args) > 0 {
			m := args[0].Get("message")
			if !m.IsUndefined() && !m.IsNull() {
				msg = m.String()
			}
		}
		if msg == "" {
			msg = "worker error"
		}
		onErr(msg)
		return nil
	})
	h.worker.Set("onmessage", h.onMessage)
	h.worker.Set("onerror", h.onError)
	return h
}

func (h *workerHandle) post(msg string) {
	if h == nil || h.dead || h.worker.IsUndefined() || h.worker.IsNull() {
		return
	}
	h.worker.Call("postMessage", msg)
}

func (h *workerHandle) terminate() {
	if h == nil || h.dead {
		return
	}
	h.dead = true
	if !h.worker.IsUndefined() && !h.worker.IsNull() {
		h.worker.Call("terminate")
	}
	if h.onMessage.Truthy() {
		h.onMessage.Release()
	}
	if h.onError.Truthy() {
		h.onError.Release()
	}
	if h.blobURL != "" {
		js.Global().Get("URL").Call("revokeObjectURL", h.blobURL)
	}
}

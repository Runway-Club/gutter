---
title: Worker
parent: Widgets
nav_order: 62
---

# `Worker`
{: .no_toc }

Offload heavy work to a Web Worker so the main UI thread stays responsive. The recommended pattern is `gutter.NewWorkerTask`: write the handler as plain Go in your main binary; the framework spawns a worker that reloads the same WASM with a flag set, dispatches to your handler, and pipes messages in and out as strings.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Worker struct {
    Task        gutter.WorkerTask
    WASM        string
    WASMExecURL string
    ScriptURL   string
    Inline      string
    Message     string
    Builder     func(WorkerSnapshot) gutter.Widget
}

type WorkerSnapshot struct {
    Pending bool
    Message string
    Error   string
    Post    func(string)
}
```

Source priority (top wins):

1. `Task` — a `gutter.WorkerTask` from `gutter.NewWorkerTask("name", handler)`. Recommended.
2. `WASM` — path to a separate Go program built with `gutter.RunWorker`.
3. `ScriptURL` — classic JS worker file.
4. `Inline` — raw JS source.

`Builder` is invoked on every rebuild with the latest snapshot. `Post` sends a follow-up message; call it from any handler.

---

## Inline Go task (recommended)

```go
// Top-level — runs before main(), so the worker bootstrap can dispatch.
var reverseTask = gutter.NewWorkerTask("reverse", func(msg string) string {
    runes := []rune(msg)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
})

// In your widget tree:
widgets.Worker{
    Task:    reverseTask,
    Message: input,
    Builder: func(snap widgets.WorkerSnapshot) gutter.Widget {
        if snap.Pending  { return widgets.Body{Text: "(reversing…)"} }
        if snap.Error != "" { return widgets.Body{Text: "Error: " + snap.Error} }
        return widgets.Body{Text: "Reversed: " + snap.Message}
    },
}
```

The worker bootstrap reloads `app.wasm` with `self.__GUTTER_WORKER_TASK = "reverse"` set. `gutter.RunApp` checks this on startup and dispatches to the registered handler instead of mounting the UI.

---

## Sending follow-up messages

`WorkerSnapshot.Post` lets you push a new message at any time:

```go
widgets.Button{
    Label: "Compute",
    OnPressed: func() { snap.Post(currentInput) },
}
```

`Pending` flips to true until the worker replies.

---

## Notes

- One Worker widget owns one worker instance for its lifetime. To restart with new source, wrap the widget in [`WithKey`](withkey.html) and change the key.
- Messages between Go and the worker are strings. For structured data, serialize (JSON is conventional) and parse on both ends.
- Cross-platform: the implementation lives in `worker_wasm.go` / `worker_stub.go`. The host stub is inert.

---

## See also

- [State Management](../state-management.html) — for `SetState` patterns that pair with workers.

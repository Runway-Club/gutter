---
title: AsyncBuilder
parent: Widgets
nav_order: 71
---

# `AsyncBuilder[T]`
{: .no_toc }

Run a `func(ctx) (T, error)` in a goroutine on mount and rebuild with an `AsyncSnapshot[T]` once it returns. The context is canceled when the widget unmounts.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type AsyncState int
const ( AsyncPending AsyncState = iota; AsyncDone; AsyncFailed )

type AsyncSnapshot[T any] struct {
    State AsyncState
    Data  T
    Error error
}

type AsyncBuilder[T any] struct {
    Load    func(ctx context.Context) (T, error)
    Builder func(ctx *gutter.BuildContext, snapshot AsyncSnapshot[T]) gutter.Widget
}
```

`Load` is called exactly once per mount. The Go function value can't be compared, so the framework can't detect when `Load` itself changed across a parent rebuild — wrap in [`WithKey`](withkey.html) with a key derived from your inputs to force a remount.

---

## Basic usage

```go
widgets.AsyncBuilder[User]{
    Load: func(ctx context.Context) (User, error) {
        return fetchUser(ctx, id)
    },
    Builder: func(_ *gutter.BuildContext, snap widgets.AsyncSnapshot[User]) gutter.Widget {
        switch snap.State {
        case widgets.AsyncPending:
            return widgets.Body{Text: "Loading…"}
        case widgets.AsyncFailed:
            return widgets.Body{Text: "Error: " + snap.Error.Error()}
        }
        return widgets.Heading{Level: widgets.H3, Text: snap.Data.Name}
    },
}
```

---

## Refetching on input change

```go
widgets.WithKey{
    Key: userID,           // changing this remounts and reruns Load
    Child: widgets.AsyncBuilder[User]{
        Load:    func(ctx context.Context) (User, error) { return fetchUser(ctx, userID) },
        Builder: /* … */,
    },
}
```

---

## Notes

- `Load` runs on the WASM single-threaded event loop; long CPU-bound work still blocks the UI. For CPU work, use [`Worker`](worker.html).
- The context passed to `Load` is canceled when the widget unmounts — wire it into HTTP requests, channels, etc.

---

## See also

- [Worker](worker.html) — for CPU-bound async work.
- [WithKey](withkey.html) — forcing re-execution.

---
title: ObserverBuilder
parent: Widgets
nav_order: 70
---

# `ObserverBuilder[T]`
{: .no_toc }

Rebuilds a leaf subtree whenever a `gutter.Listenable[T]` fires. The canonical way to bind a small piece of UI to an observable value without going through `SetState` on an ancestor.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type ObserverBuilder[T any] struct {
    Source  gutter.Listenable[T]
    Builder func(ctx *gutter.BuildContext, value T) gutter.Widget
}
```

| Field    | What it does                                                              |
| -------- | ------------------------------------------------------------------------- |
| `Source` | Any `Listenable[T]` — a `Notifier`, a `Router`'s current path, etc.       |
| `Builder`| Renders the subtree from the current value.                               |

The builder is invoked on mount, on every `Source` fire, and on parent rebuilds (the builder closure may have changed). Subscription is swapped on `DidUpdateWidget` when an ancestor replaces `Source`.

---

## Basic usage

```go
counter := gutter.NewNotifier(0)

widgets.ObserverBuilder[int]{
    Source: counter,
    Builder: func(_ *gutter.BuildContext, v int) gutter.Widget {
        return widgets.Heading{Level: widgets.H2, Text: fmt.Sprintf("Count: %d", v)}
    },
}
```

To increment from elsewhere:

```go
widgets.Button{
    Label:     "+",
    OnPressed: func() { counter.Update(func(v int) int { return v + 1 }) },
}
```

The button doesn't need to know about the heading — it just bumps the notifier.

---

## `Notifier[T]`

```go
type Notifier[T any] // exposes Value(), Set(T), Update(fn), Listen(fn) (cancel func())
```

Construct with `gutter.NewNotifier(initial)`. `Set` replaces, `Update` mutates, both fire listeners with the new value. The listener `cancel` is idempotent.

`Listenable[T]` is the read-side interface — anything implementing `Value()`/`Listen()` works as `Source`. The framework's `Router` exposes itself as `Listenable[string]` so the current path can be observed directly.

---

## Why prefer ObserverBuilder over SetState

`SetState` on a high-up State rebuilds the entire subtree below — fine for small apps, expensive for big ones. `Notifier + ObserverBuilder` lets you target the exact subtree that depends on the value: only that builder re-runs.

```go
// The Notifier sits at the root, but only the badge re-renders on changes.
unread := gutter.NewNotifier(0)

widgets.AppBar{
    Title: "Inbox",
    Actions: []gutter.Widget{
        widgets.ObserverBuilder[int]{
            Source: unread,
            Builder: func(_ *gutter.BuildContext, n int) gutter.Widget {
                return widgets.Badge{
                    Variant: widgets.BadgeNeutral,
                    Text:    fmt.Sprintf("%d", n),
                }
            },
        },
    },
}
```

---

## See also

- [AsyncBuilder](asyncbuilder.html) — same pattern for async loads.
- [State Management](../state-management.html) — when to pick `SetState` vs. `Notifier`.

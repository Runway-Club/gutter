---
title: State Management
nav_order: 4
---

# State Management
{: .no_toc }

How to add mutable state to a widget — `StatefulWidget`, `StateObject`, `SetState`, lifecycle hooks, and the reconciliation rules you need to know.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## The mental model

Gutter follows Flutter's split:

- A **Widget** is an immutable description. It's cheap, it's a value, and the framework throws it away after each Build.
- A **State** is a long-lived object the framework keeps around across rebuilds. Mutate it, then call `SetState` to ask the framework to re-Build.

Two widgets that look identical can be in different States, and that's fine — the State belongs to the Element, not to the Widget.

---

## Writing a StatefulWidget

The contract is one method:

```go
type StatefulWidget interface {
    CreateState() State
}
```

So the smallest possible StatefulWidget is:

```go
type Counter struct{}

func (Counter) CreateState() gutter.State { return &counterState{} }

type counterState struct {
    gutter.StateObject  // embed by value — gives you SetState
    count int           // your fields
}

func (s *counterState) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Button{
        Label:     fmt.Sprintf("Count: %d", s.count),
        OnPressed: func() { s.SetState(func() { s.count++ }) },
    }
}
```

Three things to internalize:

1. **`CreateState` returns a `gutter.State`** — any value with a `Build(ctx) Widget` method qualifies. In practice you return `&yourState{}`.
2. **Embed `gutter.StateObject` by value.** It carries the framework's element handle so `SetState` works.
3. **`Build` must be on the pointer receiver** (`*counterState`, not `counterState`). Mutating `s.count` on a value receiver wouldn't stick.

The widget itself (`Counter`) stays empty in this example. If your widget needs configuration from the parent — `MyChart{Data: pts}`, `TodoItem{Todo: t}` — put those fields on the widget, not on the State. The State reads them via the widget reference it can keep on itself if needed.

---

## `SetState`

```go
func (s *StateObject) SetState(fn func()) {
    fn()
    if s.elem != nil {
        s.elem.rebuild()
    }
}
```

Two steps: run `fn` (where you mutate state), then ask the framework to rebuild **only this State's subtree**.

> **Why pass a function rather than calling SetState empty?** Convention. It makes "what changed" easy to spot in code review, and it leaves room for future async/batched semantics. Today you could ignore the function and mutate state inline, but don't — write the closure so the call site reads as "between these braces, my state changed."

`SetState` is **synchronous and unbatched** today. Each call rebuilds the subtree synchronously and immediately:

```go
s.SetState(func() { s.a = 1 })
s.SetState(func() { s.b = 2 })
// → two full rebuilds. Coalesce manually:
s.SetState(func() { s.a = 1; s.b = 2 })
// → one rebuild.
```

This is a known limitation. Until the framework adds batching, the rule is: **make one SetState call per logical state transition**.

---

## Lifecycle hooks

Two optional interfaces let State observe its own lifecycle:

```go
type StateInitializer interface { InitState() }
type StateDisposer interface   { Dispose() }
```

Implement them on your State struct.

### `InitState`

Called **once**, after `CreateState`, before the first `Build`. The framework has already bound the element by this point, so `SetState` is safe — though calling it inside `InitState` is unusual.

Use `InitState` for one-shot setup: opening a channel, kicking off a background goroutine, fetching initial data, subscribing to an event source.

```go
type clockState struct {
    gutter.StateObject
    now    time.Time
    stop   chan struct{}
}

func (s *clockState) InitState() {
    s.stop = make(chan struct{})
    go func() {
        t := time.NewTicker(time.Second)
        defer t.Stop()
        for {
            select {
            case <-t.C:
                s.SetState(func() { s.now = time.Now() })
            case <-s.stop:
                return
            }
        }
    }()
}

func (s *clockState) Dispose() { close(s.stop) }
```

### `Dispose`

Called when the State's Element is unmounted — the parent decided this child is gone for good (not just re-rendered).

Use `Dispose` to release whatever `InitState` (or any rebuild) acquired: close channels, cancel contexts, unsubscribe listeners, stop tickers. If you forget, a long-running goroutine will keep calling `SetState` on a State whose Element is gone, and the call becomes a no-op — but the goroutine leaks.

---

## When does an Element get unmounted?

Three cases:

1. **The parent removed it.** `Column{Children: []gutter.Widget{a, b}}` becomes `Column{Children: []gutter.Widget{a}}` — `b`'s Element is unmounted.
2. **A reconciliation mismatch.** The new widget at the same slot is a different Go type (or a different key), so the old Element can't be reused. Old unmounted, new mounted.
3. **The whole subtree's ancestor was unmounted.** Recursive.

In all three cases, `Dispose` runs.

What *doesn't* trigger unmount is a normal rebuild where the type and key match — that's an `update`, and the State stays alive.

---

## Keyed reconciliation

Without keys, the reconciler matches list children **positionally** among siblings of the same Go type. That works fine for stable lists, but it's wrong for anything that can change order:

```go
// You have:
[TodoItem{ID: 1}, TodoItem{ID: 2}, TodoItem{ID: 3}]

// User deletes the first one:
[TodoItem{ID: 2}, TodoItem{ID: 3}]

// Without keys, the reconciler reuses:
//   slot 0's Element → now shows ID 2's data (was ID 1)
//   slot 1's Element → now shows ID 3's data (was ID 2)
//   slot 2's Element → unmounted (was ID 3)
//
// If TodoItem holds State — e.g. an "is editing" boolean — that State just got
// assigned to the wrong todo. The user's edit on row 1 is now happening on
// what looks like row 2.
```

The fix is to **key** each child by something stable across reorders — usually the item's ID:

```go
widgets.Column{
    Children: []gutter.Widget{
        widgets.WithKey{Key: 1, Child: TodoItem{Todo: todos[0]}},
        widgets.WithKey{Key: 2, Child: TodoItem{Todo: todos[1]}},
        widgets.WithKey{Key: 3, Child: TodoItem{Todo: todos[2]}},
    },
}
```

Now after the delete, the reconciler sees Keys `[2, 3]`, matches them to the old Elements that already had Keys `[2, 3]`, and unmounts only the one with Key `1`. The State on `2` and `3` is preserved.

If you control the widget, prefer implementing `gutter.Keyed` directly:

```go
type TodoItem struct{ Todo Todo }
func (t TodoItem) WidgetKey() any { return t.Todo.ID }
```

`WithKey` is the wrapper for cases where you don't own the widget.

**Rule of thumb:** key any list whose children carry their own State, or whose order can change.

---

## Pushing state down

The whole point of subtree rebuilds is that small State means small rebuilds. The lower in the tree the State lives, the less work each SetState does, and the less risk of unrelated subtrees (focused inputs, in-flight animations) getting torn down.

A common pattern: lift state up only as far as it has to go for **all the widgets that read or mutate it**, and no further.

```go
// BAD: count lives at the app root, so typing in an unrelated <input>
// rebuilds the whole app every keystroke.
type App struct{}
func (App) CreateState() gutter.State { return &appState{} }

type appState struct {
    gutter.StateObject
    count   int
    search  string
}

// BETTER: split into a Counter widget and a SearchBox widget, each with
// its own State. Typing in the search box rebuilds only the search box.
```

---

## Two-way binding for inputs

`Input.OnChanged` fires on every keystroke (DOM `input` event). The canonical pattern is:

```go
type formState struct {
    gutter.StateObject
    name string
}

func (s *formState) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Input{
        Value:       s.name,
        Placeholder: "Your name",
        OnChanged:   func(v string) { s.SetState(func() { s.name = v }) },
    }
}
```

The reconciler updates the `<input>` element in place rather than swapping it out, so **focus is preserved across rebuilds**. The user can keep typing.

---

## Beyond `SetState` — `Notifier` and `ObserverBuilder`

`SetState` is the right tool when the state and the widget that displays it sit close together. When state needs to cross subtrees — a header badge listening to a counter buried in the body, a popup open/closed flag driven by a far-away button — lift the value into a `Notifier[T]` and observe it with `ObserverBuilder[T]`.

```go
type Notifier[T any] // exposes Value(), Set(T), Update(fn), Listen(fn) (cancel func())
```

Construct with `gutter.NewNotifier(initial)`. Pass the pointer to whichever widgets need to read or write it.

```go
counter := gutter.NewNotifier(0)

widgets.Column{
    Children: []gutter.Widget{
        widgets.ObserverBuilder[int]{
            Source: counter,
            Builder: func(_ *gutter.BuildContext, v int) gutter.Widget {
                return widgets.Heading{Level: widgets.H2, Text: fmt.Sprintf("Count: %d", v)}
            },
        },
        widgets.Button{
            Label:     "+",
            OnPressed: func() { counter.Update(func(v int) int { return v + 1 }) },
        },
    },
}
```

Only the `ObserverBuilder`'s subtree rebuilds when the counter changes — no `SetState` on a high ancestor, no full-tree rebuild.

`Listenable[T]` is the read-side interface — anything implementing `Value()` and `Listen()` works as a source. Gutter's `Router` exposes itself as a `Listenable[string]` so widgets can react to navigation.

See [ObserverBuilder](widgets/observerbuilder.html) for the full pattern, and the overlay widgets ([Popup](widgets/popup.html), [Drawer](widgets/drawer.html), [BottomSheet](widgets/bottomsheet.html)) which take a `Listenable[bool]` for visibility.

---

## Async work — `AsyncBuilder`

For one-shot async loads (fetch a user, read a file from the server), [`AsyncBuilder[T]`](widgets/asyncbuilder.html) runs a `func(ctx) (T, error)` in a goroutine on mount and rebuilds with an `AsyncSnapshot` once it returns:

```go
widgets.AsyncBuilder[User]{
    Load: func(ctx context.Context) (User, error) { return fetchUser(ctx, id) },
    Builder: func(_ *gutter.BuildContext, snap widgets.AsyncSnapshot[User]) gutter.Widget {
        switch snap.State {
        case widgets.AsyncPending: return widgets.Body{Text: "Loading…"}
        case widgets.AsyncFailed:  return widgets.Body{Text: snap.Error.Error()}
        }
        return widgets.Heading{Level: widgets.H3, Text: snap.Data.Name}
    },
}
```

The context is canceled when the widget unmounts, so HTTP clients and channels listening on `ctx.Done()` clean up automatically.

For CPU-heavy work, use [`Worker`](widgets/worker.html) instead — `AsyncBuilder` runs on the WASM single-threaded event loop and would block the UI.

---

## When to reach for what

| You want to…                                                     | Pattern                                                    |
| ---------------------------------------------------------------- | ---------------------------------------------------------- |
| Mutate a value owned by this widget                              | `StatefulWidget` + `SetState`                              |
| Share a value with a non-descendant widget                       | `gutter.Notifier[T]` + [`ObserverBuilder[T]`](widgets/observerbuilder.html) |
| Show a `Loading… / Done / Error` for an async fetch              | [`AsyncBuilder[T]`](widgets/asyncbuilder.html)             |
| Animate a value between two extremes                             | [`AnimationController`](widgets/animation.html)            |
| Toggle an overlay (popup / drawer / sheet)                       | `Notifier[bool]` consumed by the overlay widget directly   |
| Run CPU-bound work without blocking the UI                       | [`Worker`](widgets/worker.html)                            |

---

## Cheat sheet

| You want to…                            | Do this                                                                                  |
| --------------------------------------- | ---------------------------------------------------------------------------------------- |
| Add a counter / toggle / form value     | Make a `StatefulWidget`, embed `StateObject`, call `SetState`.                           |
| Run code once when the widget mounts    | Implement `InitState()` on your State.                                                   |
| Clean up when the widget goes away      | Implement `Dispose()` on your State.                                                     |
| Preserve State across list reorder      | Implement `Keyed` on the widget, or wrap it in `widgets.WithKey`.                        |
| Coalesce multiple changes into one rebuild | Put all the mutations inside a single `SetState(func() { … })` call.                  |
| Share state between sibling widgets     | Lift the State into the nearest common ancestor StatefulWidget, OR use a `Notifier[T]` shared by both. |
| Update an input without losing focus    | Just `SetState` — the reconciler updates the `<input>` in place, the DOM `value` property is synced caret-preserving. |

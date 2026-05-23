---
title: Architecture
nav_order: 3
---

# Architecture
{: .no_toc }

How the framework actually works under the hood — the three-tier widget model, the persistent Element tree, and the reconciliation algorithm that diffs them.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## The three-tier widget model

A `Widget` in Gutter is literally `type Widget = any`. There is no required base interface. Instead, the framework dispatches on whichever of three interfaces your value implements:

```go
type HostWidget interface {
    Host() *Host
}

type StatelessWidget interface {
    Build(ctx *BuildContext) Widget
}

type StatefulWidget interface {
    CreateState() State
}
```

### HostWidget — a leaf that maps to one DOM element

`HostWidget` is what the reconciler ultimately materializes. Every node that becomes a real DOM element implements it. The catalog's primitives (`Container`, `Column`, `Row`, `Center`, `Padding`, `SizedBox`, `Text`, `Styled`) are HostWidgets.

```go
type Host struct {
    Tag      string                       // "div", "button", "span", …
    Text     string                       // textContent, if any
    Attrs    map[string]string            // id, href, type, value, …
    Style    map[string]string            // inline CSS
    Events   map[string]func(Event)       // listener functions
    Children []Widget                     // sub-widgets, mounted recursively
}
```

The framework — not the widget — recursively mounts `Children`. You return a tree; the runtime walks it.

### StatelessWidget — composes by returning another widget

A StatelessWidget exists only to produce another widget. It has no DOM of its own; its Element delegates `dom()` to its single child.

```go
type Hello struct{ Name string }

func (h Hello) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Heading{Level: widgets.H2, Text: "Hello, " + h.Name}
}
```

Most themed widgets — `Heading`, `Body`, `Button`, `Card`, `Surface`, `Input`, `Badge`, `AppBar`, `Scaffold` — are StatelessWidgets. They read `ctx.Theme`, resolve the right tokens, and return a `Styled` (or a tree of primitives).

### StatefulWidget — owns mutable State across rebuilds

A StatefulWidget itself stays immutable; the **state** lives in a separate object that survives rebuilds.

```go
type Counter struct{}
func (Counter) CreateState() gutter.State { return &counterState{} }

type counterState struct {
    gutter.StateObject
    count int
}

func (s *counterState) Build(ctx *gutter.BuildContext) gutter.Widget { /* … */ }
```

When the framework mounts a `StatefulWidget`, it calls `CreateState()` once and stashes the returned `State`. On every rebuild, it calls `state.Build(ctx)` — never `CreateState` again. Your local `count` persists.

See [State Management](state-management.html) for the full story.

---

## The persistent Element tree

Widgets are immutable descriptions. The runtime maintains a parallel, **persistent** tree of `Element`s — one per widget — that owns the actual DOM nodes, event listeners, and (for StatefulWidgets) the State.

| Element type        | Owns                                                          | DOM?                       |
| ------------------- | ------------------------------------------------------------- | -------------------------- |
| `hostElement`       | one DOM node + a list of child Elements + `js.Func` listeners | yes (the element itself)   |
| `statelessElement`  | one child Element                                             | no — delegates to child    |
| `statefulElement`   | a `State` + one child Element                                 | no — delegates to child    |

The `Element` interface is the seam between platform-independent widget types and the WASM runtime:

```go
mount(parent, before, ctx)   // create DOM, insert into parent
update(newWidget, ctx)       // diff newWidget against current, mutate DOM in place
unmount()                    // remove DOM, release listeners, recurse, call State.Dispose
dom()                        // return root DOM node owned by this element (or descendants)
```

When the tree changes — because state was set, or an ancestor rebuilt — the runtime calls `update` on Elements that can be reused, `mount` on new ones, and `unmount` on the rest.

---

## Reconciliation

Reconciliation is how Gutter turns "here's the new widget tree" into the smallest possible set of DOM mutations.

### Single-child case

```go
reconcile(parent, oldEl, newW, ctx)
```

If `canUpdate(oldEl, newW)` — same Go type **and** same key — the existing Element is kept and asked to update itself against the new widget. Otherwise, a fresh Element is mounted at `oldEl`'s DOM position, and `oldEl` is unmounted.

### List case (the interesting one)

```go
reconcileChildren(parent, oldChildren, newWidgets, ctx)
```

This is what the framework runs for every list of children — `Column.Children`, `Row.Children`, `Styled.Children`, etc. The algorithm:

1. Build `map[key]index` over the keyed old children.
2. Walk `newWidgets`. For each one:
   - If keyed, look it up in the map. If the type matches, reuse that old Element.
   - Otherwise, take the next unused old child of the same Go type, positionally.
3. Unmount unmatched old children.
4. Walk the resulting list **backwards** and `insertBefore` each Element, using the previously-placed sibling as the anchor. This positions newly mounted nodes and moves reused nodes in a single O(n) pass.

This is the same algorithm React uses for keyed-vs-positional reconciliation.

### Keys

```go
type Keyed interface { WidgetKey() any }
```

A widget that implements `Keyed` participates in keyed matching. The catalog ships a generic wrapper:

```go
widgets.WithKey{Key: todo.ID, Child: TodoItem{Todo: todo}}
```

Without keys, unkeyed siblings of the same Go type are matched positionally. That's fine for stable lists — but for lists that can be **reordered, inserted in the middle, or deleted from the middle**, you almost always want keys, because positional matching will reuse the wrong Element and you'll see "the wrong item animated" or "the focus jumped to a different input."

Rule of thumb: if the list can change shape, key it.

---

## SetState and subtree rebuilds

`StateObject.SetState(fn)` does two things, in order:

1. Runs `fn()` so you can mutate your state.
2. Calls `s.elem.rebuild()` if the State has been mounted.

`rebuild()` is the punchline of the whole architecture. It is implemented on `statefulElement` as:

```go
func (e *statefulElement) rebuild() {
    newChild := e.state.Build(e.ctx)
    e.child = reconcile(e.parent, e.child, newChild, e.ctx)
}
```

**Only this Element's subtree is rebuilt.** The rest of the tree — siblings, ancestors, other StatefulWidgets, focused inputs, animations, scroll positions — is untouched.

This is the same model Flutter and React use, and it's why the canonical Gutter pattern is to push state down to the smallest possible StatefulWidget that owns it. Lift state up when you must, but keep the rebuild surface small.

---

## Theming via `BuildContext`

`BuildContext` is threaded through every `Build` and `mount`. Today it carries one thing:

```go
type BuildContext struct {
    Theme *themes.Theme
}
```

The same `*BuildContext` instance is shared by every Build call in the app. There is **one BuildContext per app**, not one per widget. This is intentional: it lets `Scaffold` set `ctx.Theme = s.Theme` once and have the change propagate to every descendant — a "theme provider" without an InheritedWidget mechanism.

If you need per-subtree theming later, the path forward is to add an InheritedWidget-style propagation API on top of `BuildContext`.

---

## Build tags and the WASM boundary

Anything that imports `syscall/js` lives in a `*_wasm.go` file with `//go:build js && wasm`. Today that's:

- `app_wasm.go` — `RunApp`
- `element_wasm.go` — the Element tree, reconciler, DOM glue
- `title_wasm.go` — `SetTitle` (writes `document.title`)

For each `*_wasm.go` there's typically a `*_stub.go` with the inverse build tag (`!js || !wasm`) that defines a panicking stub. The stubs let user `package main` code compile and pass `go vet` on the host, even though calling `RunApp` outside of `GOOS=js GOARCH=wasm` would panic.

This is why every Gutter project has two `go build` paths that matter:

```sh
go build ./...                              # editor / vet on host
GOOS=js GOARCH=wasm go build -o app.wasm .  # the actual artifact
```

The `gutter` CLI handles the second one for you.

---

## Where to look in the source

| File                                                                                                                | What's in it                                                              |
| ------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| [`gutter.go`](https://github.com/Runway-Club/gutter/blob/main/gutter.go)                                            | `Widget`, `HostWidget`, `StatelessWidget`, `StatefulWidget`, `Host`, `Event`, `Keyed`. |
| [`state.go`](https://github.com/Runway-Club/gutter/blob/main/state.go)                                              | `State`, `StateInitializer`, `StateDisposer`, `StateObject`.              |
| [`context.go`](https://github.com/Runway-Club/gutter/blob/main/context.go)                                          | `BuildContext`.                                                           |
| [`options.go`](https://github.com/Runway-Club/gutter/blob/main/options.go)                                          | `Option`, `WithTheme`, `WithSelector`.                                    |
| [`app_wasm.go`](https://github.com/Runway-Club/gutter/blob/main/app_wasm.go)                                        | `RunApp`.                                                                 |
| [`element_wasm.go`](https://github.com/Runway-Club/gutter/blob/main/element_wasm.go)                                | The Element tree and reconciler.                                          |
| [`widgets/internal.go`](https://github.com/Runway-Club/gutter/blob/main/widgets/internal.go)                        | `activeTheme`, `applySpec`, `styleFromSpec` — how themed widgets resolve theme tokens. |

---

## Known limitations

- **`SetState` is synchronous and unbatched.** Two back-to-back calls rebuild twice. Coalesce manually if you care.
- **Tag stability is assumed for HostWidgets.** `canUpdate` checks Go type but not the rendered `Host.Tag`. If a single HostWidget type produced different tags based on its fields, the framework would try to update a `<div>` into a `<span>` via attribute diffing, which produces a wrong DOM. None of the built-in widgets do this; if you write your own HostWidget, keep its tag stable per Go type.
- **No InheritedWidget yet.** Per-subtree theming, dependency injection, locale propagation — all of these would benefit from a proper inherited-data mechanism. Today, Scaffold's "set `ctx.Theme` and let the whole tree see it" is the only path.
- **No async, animation, router, image, SSR.** Roadmap items.

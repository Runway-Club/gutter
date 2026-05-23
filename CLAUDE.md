# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Gutter is a Go library for building web applications declaratively, inspired by Flutter. It compiles to WebAssembly and drives the DOM directly via `syscall/js`. The repo also ships a developer CLI in `cmd/gutter` that scaffolds projects and serves WASM builds.

## Package layout

- `github.com/Runway-Club/gutter` — framework core: `Widget`, `Host`, `Event`, `State`, `StateObject`, `BuildContext` (carries `Theme`), `Keyed`, `RunApp`, `Option`/`WithTheme`/`WithSelector`, `SetTitle` (cross-platform document title helper — WASM impl + host stub), and the runtime (Element tree, reconciler).
- `github.com/Runway-Club/gutter/themes` — theme data: `Theme`, `Colors`, `Typography`, `Rounded`, `Spacing`, `Components`, plus the `Apple`, `Meta`, and `Neutral` presets. **No runtime — pure data structs.** `Apple` is the framework default (set in `options.go`).
- `github.com/Runway-Club/gutter/widgets` — the single widget catalog. Two flavors live here side by side:
  - **App shell + themed** (StatelessWidgets that read `ctx.Theme`): `Scaffold` (the recommended root — title, theme, app bar, body, footer), `AppBar`, `Heading`, `Body`, `Caption`, `Link`, `Button`, `Card`, `Surface`, `Input`, `Badge`.
  - **Primitive / layout** (no theme dependency): `Text`, `Container`, `Styled`, `Column`, `Row`, `Center`, `Padding`, `SizedBox`, `WithKey`.
  Apps should reach for `Scaffold` first; primitives are escape hatches.
- `github.com/Runway-Club/gutter/cmd/gutter` — developer CLI.
- `github.com/Runway-Club/gutter/examples/...` — runnable examples (counter, showcase).
- `github.com/Runway-Club/gutter/theme_specs` — the YAML+prose design specs (Apple, Meta) the presets are extracted from. **Source of truth for theme values.**

Dependency direction (one-way): `themes` imports nothing in the module · `gutter` imports `themes` (for `BuildContext.Theme` and the default in `options.go`) · `widgets` imports `gutter` and `themes`. There is intentionally no `widgets/themed` subpackage — keeping the catalog flat means app code doesn't have to choose between two parallel APIs.

## Architecture

Gutter has a three-tier widget model layered over a persistent Element tree, the way Flutter and React do it.

### Widgets

A `Widget` is just `any`; concrete types are dispatched on by a type switch.

- **HostWidget** (`Host() *Host`) — leaf that maps to one DOM element. The standard catalog ships these in `widgets/`.
- **StatelessWidget** (`Build(ctx) Widget`) — composes by returning another widget.
- **StatefulWidget** (`CreateState() State`) — creates a `State`. `State.Build(ctx) Widget` is invoked on every rebuild.

`Host` carries `Tag`, `Text`, `Attrs`, `Style`, `Events`, and `Children []Widget`. The framework — not the widget — recursively mounts children.

### Element tree

`element_wasm.go` defines a persistent Element for every Widget in the tree:

- `hostElement` owns a DOM node and a list of child Elements. Owns the `js.Func` event listeners.
- `statelessElement` owns a single child Element. Has no DOM of its own; `dom()` delegates to the child.
- `statefulElement` owns the `State` plus a single child Element. Like statelessElement it is DOM-transparent.

The Element interface is the seam between platform-independent widget definitions and the WASM runtime:

```
mount(parent, before, ctx)   create DOM, insert into parent (insertBefore semantics)
update(newWidget, ctx)       diff newWidget against current, mutate DOM in place
unmount()                    remove DOM, release listeners, recurse, call State.Dispose
dom()                        root DOM node owned by this element (or descendants)
```

### Reconciliation

`reconcile(parent, oldEl, newW, ctx)` handles the single-child case: if `canUpdate(oldEl, newW)` (same Go type + same key), it calls `oldEl.update`; otherwise it mounts a fresh element at `oldEl`'s DOM position and unmounts the old one.

`reconcileChildren(parent, oldChildren, newWidgets, ctx)` handles lists:

1. Build a `map[key]index` over keyed old children.
2. For each new widget, match by key first; if unkeyed, take the next unused old child with the same Go type.
3. Unmount unmatched old children.
4. Walk the result list **backwards** and `insertBefore` each element using the previously-positioned sibling as the anchor — this places newly mounted nodes and moves reused nodes in a single O(n) pass.

This is the algorithm React uses. Without keys, reordering can't be detected and any unkeyed sibling of the same type is fair game for reuse.

### Keys

`Keyed` is implemented by widgets that participate in keyed reconciliation:

```go
type Keyed interface {
    WidgetKey() any
}
```

`widgets.WithKey{Key: ..., Child: ...}` (in the widgets package) is the canonical wrapper for keying a widget that doesn't have a key field of its own. Internally `WithKey` is a `StatelessWidget` whose `Build` returns `Child`; it shows up in the Element tree as one extra `statelessElement` layer.

### State persistence and SetState

`StateObject` (embed by value, return state by pointer) gives `SetState`. On mount, `statefulElement.mount` checks whether the State satisfies `elementBinder` (provided automatically by `StateObject`) and injects itself via `bindElement(self)`. The State now holds an opaque `stateElement` handle.

`SetState(fn)` mutates state then calls `s.elem.rebuild()`. `statefulElement.rebuild` calls `state.Build(ctx)` and reconciles the result against its current child — **only the subtree** owned by this element is rebuilt, not the whole tree. Other siblings, other StatefulWidgets, animations, focused inputs, none of them are touched.

### Lifecycle hooks

- `StateInitializer.InitState()` — called once after `CreateState`, before first `Build`.
- `StateDisposer.Dispose()` — called when the `statefulElement` is unmounted.

### Theming

Themes are passed in either via `gutter.WithTheme(themes.Meta)` to `RunApp` OR via `widgets.Scaffold{Theme: themes.Meta}` (recommended — Scaffold is also where the app's title and chrome live). Calling `RunApp(root)` with no options and no Scaffold override defaults to `themes.Apple`. The chosen theme is stored on `BuildContext.Theme` and is the **same instance for the whole app** (one BuildContext, shared by every Build call). There is no per-subtree theme override yet — that would need an InheritedWidget mechanism. Scaffold sets `ctx.Theme` directly during its Build; because the same `*BuildContext` is threaded through every descendant, the mutation propagates.

Themed widgets are all `StatelessWidget`s. They read `ctx.Theme` via the package-private `activeTheme(ctx)` (in `widgets/internal.go`), pull the right `TextSpec` / `ButtonStyle` / `CardStyle` etc. out of `theme.Components` or `theme.Typography`, render via `widgets.Styled` with the resolved CSS, and never inline a hex literal. **App code should not need to write CSS or reference theme tokens directly.** When `ctx` is nil or `ctx.Theme` is nil (e.g. in unit tests), `activeTheme` falls back to `themes.Apple` so widgets still render.

**Default font is Lexend.** Every built-in theme's TextSpecs lead with `Lexend, ...` and fall back to the brand's proprietary font (SF Pro on Apple, Optimistic VF on Meta) and then system-ui. The scaffolded `index.html` (`cmd/gutter/new.go`) and the example HTMLs preload Lexend from Google Fonts; without that `<link>`, the stack falls through to system fonts.

The presets (`themes.Apple`, `themes.Meta`) are extracted by hand from the design docs in `theme_specs/`. **When updating either, edit the corresponding spec first** — it's the source of truth — then mirror the changes into `themes/apple.go` or `themes/meta.go`. Keep the YAML→Go mapping faithful so the linter-style comments in the spec stay meaningful.

## Build tags

Anything that imports `syscall/js` lives in a `*_wasm.go` file with `//go:build js && wasm`. This currently means `app_wasm.go` and `element_wasm.go`.

`app_stub.go` (`//go:build !js || !wasm`) provides panicking stubs for `RunApp` and `Run` so that user `package main` code compiles on the host platform too — useful for editor analysis and `go vet`. The stubs panic if actually called; the binary is only meaningful when built with `GOOS=js GOARCH=wasm`.

If you add code that uses `js.Value`, put it in a `_wasm.go` file with the build tag, and keep the platform-neutral abstractions in plain files.

## Commands

```sh
go build ./...                                    # library + CLI on host
GOOS=js GOARCH=wasm go build ./...                # WASM compile of library
go vet ./... && GOOS=js GOARCH=wasm go vet ./...  # vet both targets
go build -o bin/gutter ./cmd/gutter               # build the CLI

cd examples/counter
GOOS=js GOARCH=wasm go build -o app.wasm .        # build example to WASM
go run ../../cmd/gutter run                       # build + serve on :8080
```

No tests yet; WASM tests need a browser harness which the repo doesn't wire up.

## CLI internals (`cmd/gutter`)

- `new.go` writes three templated files (`main.go`, `index.html`, `go.mod`) using `strings.ReplaceAll` rather than `text/template`. Keep it that way unless templates grow.
- `run.go` shells out to `go build` with `GOOS=js GOARCH=wasm`, then copies `wasm_exec.js` from `$GOROOT/lib/wasm/` (Go 1.24+) or `$GOROOT/misc/wasm/` (older). Both paths are tried.
- `gutter run` mounts an `http.FileServer` on the current directory and registers `.wasm → application/wasm` before serving. Don't drop that — browsers reject WASM with the wrong MIME.

## Module path and examples

The module is `github.com/Runway-Club/gutter`. `examples/counter/go.mod` uses `replace github.com/Runway-Club/gutter => ../..` so the example builds against the working copy. Mirror that for any new example.

`gutter new` does **not** emit a replace directive; users from a local checkout add one themselves. Intentional — most users will `go get` once the module is published.

## Known limitations

- No async batching of `SetState`: each call rebuilds the subtree synchronously. Two `SetState`s back-to-back rebuild twice.
- Tag stability is assumed for HostWidgets: the canUpdate check uses Go type but not the rendered `Host.Tag`. If a single HostWidget type produced different tags based on its fields, the framework would try to update a `<div>` into a `<span>` via attribute diffing, which produces a wrong DOM. None of the built-in widgets do this.
- No tests, no devtools, no portal/teleport, no SSR, no animations, no router.

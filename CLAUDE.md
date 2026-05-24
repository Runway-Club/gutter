# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

Gutter is a Go library for building web applications declaratively, inspired by Flutter. It compiles to WebAssembly and drives the DOM directly via `syscall/js`. The repo also ships a developer CLI in `cmd/gutter` that scaffolds projects and serves WASM builds.

## Package layout

- `github.com/Runway-Club/gutter` — framework core: `Widget`, `Host`, `Event`, `State`, `StateObject` (exposes `Widget()` — the framework keeps the current widget pointer fresh on mount and on every update, so State code doesn't have to copy it), `WidgetUpdater` (optional hook called when a parent rebuild swaps in a new widget instance of the same type — used by `ObserverBuilder`/`AsyncBuilder` to resubscribe), `BuildContext` (carries `Theme`), `Keyed`, `Listenable[T]`/`Notifier[T]` (observable primitive — `Value`/`Set`/`Update`/`Listen` with idempotent cancel; pair with `ObserverBuilder` for reactive subtrees), `RunApp`, `NewWorkerTask`/`WorkerTask` (register an inline worker handler — `RunApp` dispatches to it when the worker bootstrap reloads `app.wasm` with `__GUTTER_WORKER_TASK` set), `RunWorker` (lower-level worker entry, used when a worker is its own binary), `Option`/`WithTheme`/`WithSelector`, `SetTitle` (cross-platform document title helper — WASM impl + host stub), `AssetURL`/`SetAssetBase`/`AssetBaseURL` (resolve a relative asset path against a configurable base — defaults to `"assets/"`, matching what the CLI copies into `./dist/assets/`; absolute URLs and `data:` URIs pass through unchanged so widget `Asset`/`Src` fields can take either), and the runtime (Element tree, reconciler).
- `github.com/Runway-Club/gutter/themes` — theme data: `Theme`, `Colors`, `Typography`, `Rounded`, `Spacing`, `Components`, plus the `Apple`, `Meta`, and `Neutral` presets. **No runtime — pure data structs.** `Apple` is the framework default (set in `options.go`).
- `github.com/Runway-Club/gutter/widgets` — the single widget catalog. Three flavors live here side by side:
  - **App shell + themed** (StatelessWidgets that read `ctx.Theme`): `Scaffold` (the recommended root — `Title`/`Theme`/`AppBar`/`StickyAppBar`/`Body`/`Footer`; when `StickyAppBar` is true the bar is wrapped in `position:sticky; top:0; z-index:900` so it pins to the viewport while the body scrolls past — z-index sits below the 1000 overlay tier so Popup/Drawer/BottomSheet still cover it), `AppBar`, `Heading`, `Body`, `Caption`, `Link`, `Button`, `IconButton` (square Button variant rendering an `Icon` as its only content), `Card`, `Surface`, `Badge`, `Image` (HTML `<img>` with `Asset` resolved via `gutter.AssetURL` or absolute `Src`; supports `Fit` for object-fit), `Icon` (Google Material Symbols glyph; `Style: IconOutlined|IconRounded|IconSharp`, `Filled`, `Weight`, `Grade` — drives the FILL/wght/GRAD/opsz axes via font-variation-settings; the scaffolded `index.html` preloads all three stylesheets), `File` (themed file picker — `Label`/`Child` for the trigger styled like a Button, `Accept`/`Multiple`, callback receives `[]FilePick{Name, Size, MimeType, Data []byte}` with bytes pre-read via FileReader; reading is WASM-only, `file_wasm.go`/`file_stub.go` split).
  - **Input family** (all themed via `theme.Components.Input` / `theme.Colors.Primary`; controlled — declarative `Value`/`Checked`/`Selected` field is the source of truth, change callback fires on user edits, parent rebuilds with the new value): `Input` (single-line; `Type: InputText|Password|Email|Number|Tel|URL|Search|Date|Time|DateTimeLocal|Month|Week|Color` plus `Min/Max/Step/Pattern/AutoComplete/Disabled/ReadOnly/Name`), `TextArea` (multi-line; `Rows`, `Resize`, `MaxLength`), `Checkbox` (label-wrapped native checkbox themed via `accent-color`), `Switch` (custom CSS sliding pill — implemented as a styled `<button role="switch">` with a sibling thumb, since `:checked +` selectors don't work inline; controlled visually from the `Checked` field), `Slider` (`<input type="range">` with `accent-color`), `Select[T comparable]` + `SelectOption[T]` (generic dropdown; HTML option value is the slice index so any comparable T works — `Placeholder` injects a disabled initial option), `RadioGroup[T comparable]` + `RadioOption[T]` (generic group; StatefulWidget assigns a stable `name` per mount via `InitState` and an atomic counter so multiple groups don't bleed into one). The "controlled" semantics come from `propSyncHost` (in `internal.go`) — a HostWidget that exposes `OnMount`/`OnUnmount` alongside the Styled bundle so widgets can imperatively set DOM properties (`checked`, `value`) after every reconcile via `setBoolProp`/`setStringProp` (`dom_wasm.go`/`dom_stub.go`), because `applyAttrs` only sets the default-state attribute. Text-style inputs (`Input`, `TextArea`) use the caret-preserving `setStringPropIfDifferent` variant — writing `value` on every reconcile would move the caret to the end of the string on each keystroke, so we read the current DOM value first and skip the write when it already matches state.
  - **Primitive / layout** (no theme dependency): `Text`, `Container`, `Styled`, `Column`, `Row`, `Center`, `Padding`, `SizedBox`, `WithKey`, `Transform` (CSS translate/rotate/scale/skew wrapper — the zero value is the identity), `Controller[T]` + `Draggable[T]` + `DropTarget[T]` + `DragOverlay[T]` (pointer-based drag-and-drop kit; one Controller per drag domain coordinates sources and targets; Draggable wires `pointerdown` → `setPointerCapture` → window-level `pointermove`/`pointerup` listeners; DropTarget registers its DOM node with the controller and the wasm-side hit-tester picks the *smallest* containing `getBoundingClientRect` on every move so nested targets work naturally; DragOverlay renders the ghost `position: fixed; pointer-events: none` so cursor hit-tests pass through to targets; `dragdrop_wasm.go`/`dragdrop_stub.go` split), `List` (eager scrollable flex container — Column/Row plus `overflow:auto` and a bounded `Height`/`Width`; `Direction: ListVertical|ListHorizontal`; opt out with `NoScroll`), `ListBuilder` (vertical-only virtualized list — `ItemCount`, fixed `ItemHeight`, `ItemBuilder(i)`, `Height`, `Overscan` default 3; renders only the visible window into a `position:absolute` wrapper offset by `firstVisible*ItemHeight` inside a `position:relative` sizer of the full virtual height; the scroll listener is attached once via `propSyncHost.OnMount` in `list_wasm.go`/`list_stub.go` with `{passive: true}`, fires once synchronously to seed `viewportHeight` before first paint, and `SetState`s only when `firstVisible` or `viewportHeight` actually change so scroll handlers below the item-boundary stride short-circuit. **Recycling contract**: keep `ItemBuilder`'s root widget type stable across indices and do NOT key items with `WithKey` — positional matching on same Go type is what lets the reconciler update DOM in place as the window shifts; keying would force unmount+remount on every scroll. State belongs to the slot, not the data — prefer Stateless items).
  - **Overlays** (themed; visibility driven by a `Listenable[bool]` the app holds — typically a `Notifier[bool]`): `Popup` (centered modal with backdrop + fade/scale transition), `Drawer` (slide-in side panel, `Side: DrawerLeft|DrawerRight`), `BottomSheet` (slide-up bottom panel). All three render both the open and closed CSS states so transitions run in both directions, and use a sibling backdrop+content pair under a `display: contents` wrapper so backdrop clicks dismiss without sheet clicks bubbling into the handler.
  - **Imperative / lifecycle**: `Canvas` (typed 2D painter, uses `Host.OnMount`), `GestureDetector` (wraps a child with pointer/key event hooks via `display: contents`), `Worker` (offloads heavy work to a Web Worker via a builder/snapshot API; primary source mode is `Task: gutter.NewWorkerTask("name", func(msg string) string {...})` — the inline handler lives in the main app binary and the worker bootstrap reloads `app.wasm` with `self.__GUTTER_WORKER_TASK` set so `RunApp` dispatches to it instead of mounting). Cross-platform — `canvas_wasm.go`/`canvas_stub.go` and `worker_wasm.go`/`worker_stub.go` keep `syscall/js` out of host builds.
  - **Reactive / control flow**: `ObserverBuilder[T]` (subscribes to a `gutter.Listenable[T]` and rebuilds when it fires; swaps subscriptions on `DidUpdateWidget`), `AsyncBuilder[T]` (runs a `func(context.Context) (T, error)` in a goroutine, rebuilds with an `AsyncSnapshot[T]` of `{State: Pending|Done|Failed, Data, Error}`; re-invoke by wrapping in `WithKey` with a new key), `Router`/`RouterView` (path-based routing with `:param` captures, hooks browser `popstate` + `history.pushState/replaceState/back` via `router_wasm.go`/`router_stub.go`; the router itself is a `Listenable[string]` so other widgets can observe the current path), `AnimationController` + `AnimatedBuilder` (time-driven `Listenable[float64]` interpolating between `Lower` and `Upper` over `Duration` with a `Curve` — `CurveLinear/EaseIn/EaseOut/EaseInOut`; ticks at 60Hz from a goroutine via `time.NewTicker`, so it's cross-platform with no `_wasm` split; pair with `Transform` for motion).
  Apps should reach for `Scaffold` first; primitives are escape hatches.
- `github.com/Runway-Club/gutter/cmd/gutter` — developer CLI.
- `github.com/Runway-Club/gutter/community/...` — vendor-specific reusable widgets that intentionally don't belong in the core `widgets` catalog. Currently: `community/login_with_google` — Google Identity Services button. The widget loads `https://accounts.google.com/gsi/client` once per page, runs `google.accounts.id.initialize/renderButton`, bridges the callback into Go via `js.FuncOf`, and parses the returned JWT into a typed `Credential` (`Token`, `Sub`, `Email`, `EmailVerified`, `Name`, `GivenName`, `FamilyName`, `Picture`, `Issuer`, `Audience`, `Expiry`). `gsi_wasm.go`/`gsi_stub.go` split. The package also exports `GLogoDataURL` (Google "G" mark as an inline SVG data: URL) for app-side branding. Trust model: client-side JWT parse is convenience-only; production code must re-verify `Credential.Token` server-side against Google's public keys.
- `github.com/Runway-Club/gutter/examples/...` — runnable examples (counter, showcase, playground, router).
- `github.com/Runway-Club/gutter/theme_specs` — the YAML+prose design specs (Apple, Meta) the presets are extracted from. **Source of truth for theme values.**

Dependency direction (one-way): `themes` imports nothing in the module · `gutter` imports `themes` (for `BuildContext.Theme` and the default in `options.go`) · `widgets` imports `gutter` and `themes` · `community/*` imports `gutter` (and may import `widgets` if it composes themed widgets). There is intentionally no `widgets/themed` subpackage — keeping the catalog flat means app code doesn't have to choose between two parallel APIs.

## Architecture

Gutter has a three-tier widget model layered over a persistent Element tree, the way Flutter and React do it.

### Widgets

A `Widget` is just `any`; concrete types are dispatched on by a type switch.

- **HostWidget** (`Host() *Host`) — leaf that maps to one DOM element. The standard catalog ships these in `widgets/`.
- **StatelessWidget** (`Build(ctx) Widget`) — composes by returning another widget.
- **StatefulWidget** (`CreateState() State`) — creates a `State`. `State.Build(ctx) Widget` is invoked on every rebuild.

`Host` carries `Tag`, `Text`, `Attrs`, `Style`, `Events`, `Children []Widget`, and the two escape-hatch hooks `OnMount(node any)` / `OnUnmount(node any)`. The framework — not the widget — recursively mounts children. The hooks fire after the DOM is inserted (and after every update for `OnMount`) / before it is removed, with `node` being the platform-native handle (`syscall/js.Value` on WASM). They exist so widgets that need imperative DOM access (`Canvas` calling `getContext`) or external resources tied to a placeholder (`Worker`) can do so without leaking `syscall/js` into platform-neutral widget structs. `Event` was likewise extended with `X/Y` (clientX/clientY), `OffsetX/OffsetY`, and `Key`, which `GestureDetector` exposes.

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
4. Walk the result list **backwards** and `insertBefore` each element using the previously-positioned sibling as the anchor — this places newly mounted nodes and moves reused nodes in a single O(n) pass. The call is skipped when `current.nextSibling === nextDom` (the node is already where it belongs); per the DOM spec re-inserting in place is a no-op, but in practice Chrome treats it as a remove+insert that blurs and refocuses any focused descendant, which on a long page triggers the browser's `scrollIntoView` on focus restore and snaps scroll back to where the input lives. The cheap reference compare avoids that whole class of "page scrolls to top on every keystroke" bugs.

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
- `build.go` / `wasm.go` shell out to `go build` with `GOOS=js GOARCH=wasm`, bundle into `./dist/`, and copy `wasm_exec.js` from `$GOROOT/lib/wasm/` (Go 1.24+) or `$GOROOT/misc/wasm/` (older). Both paths are tried. `bundleInto` also walks `./public/` (root-level assets — favicons, robots.txt, etc.) and `./assets/` (referenced by `widgets.Image{Asset: ...}` via `gutter.AssetURL`) into `./dist/` and `./dist/assets/` respectively. Both are no-ops if the source directory is missing.
- `--tinygo` (opt-in, on `build`/`run`/`run dev`/`build deploy`) threads a `tinygo bool` through `bundleInto` → `buildWasm`/`buildWasmPkg`/`bundleWorkers`/`ensureWasmExec`. In tinygo mode the CLI runs `tinygo build -o … -target wasm` instead of `go build`, and copies `wasm_exec.js` from `$(tinygo env TINYGOROOT)/targets/` instead of GOROOT. `ensureWasmExec` always overwrites (Go and TinyGo ship incompatible runtimes), and `ensureTinygo` fails fast with an install hint if the binary is missing. **TinyGo gotcha**: its `fmt.Sscanf` panics on `%f` (float scan) rather than returning an error — don't use `Sscanf` for float parsing in widget code; use `strconv.ParseFloat` (see `parseSizePx` in `widgets/icon.go` and `parsePxFallback` in `widgets/list.go`). Default builds remain pure Go.
- The primary worker pattern needs no CLI help — inline tasks via `gutter.NewWorkerTask` ship in the same `app.wasm` and the worker bootstrap reloads that single binary. `bundleWorkers` is a legacy escape hatch for genuinely separate worker binaries: `./worker/` → `dist/worker.wasm`, `./workers/<name>/` → `dist/workers/<name>.wasm`. Use it only when you don't want the worker to pull in the whole app.
- `run.go` serves `./dist/` over HTTP and registers `.wasm → application/wasm`. Don't drop that — browsers reject WASM with the wrong MIME. `gutter run dev` adds an fsnotify watcher + a tiny `/__gutter/build` poller injected into served HTML for live reload.

## Module path and examples

The module is `github.com/Runway-Club/gutter`. `examples/counter/go.mod` uses `replace github.com/Runway-Club/gutter => ../..` so the example builds against the working copy. Mirror that for any new example.

`gutter new` does **not** emit a replace directive; users from a local checkout add one themselves. Intentional — most users will `go get` once the module is published.

## Known limitations

- No async batching of `SetState`: each call rebuilds the subtree synchronously. Two `SetState`s back-to-back rebuild twice.
- Tag stability is assumed for HostWidgets: the canUpdate check uses Go type but not the rendered `Host.Tag`. If a single HostWidget type produced different tags based on its fields, the framework would try to update a `<div>` into a `<span>` via attribute diffing, which produces a wrong DOM. None of the built-in widgets do this.
- No tests, no devtools, no portal/teleport (overlays are siblings under a `display:contents` wrapper, not in a true root portal), no SSR.
- `ListBuilder` requires a fixed `ItemHeight` — variable-height rows would need a measurement and offset cache, which the current implementation doesn't do. Horizontal virtualization isn't supported either; the eager `List` covers horizontal scroll cases.
- Form-element controlled inputs sync through DOM properties on `OnMount` (`setStringPropIfDifferent` for text inputs/textareas to preserve caret, `setBoolProp` for checkboxes/radios, `setStringProp` for sliders/selects) because `applyAttrs` only writes the default-state attribute and re-writing `value` via setAttribute on every keystroke would move the caret to the end.
- Router (`widgets.Router`) is path-only: no nested routers, no guards, no transitions, no query-parameter parsing beyond passing the raw search string through. Wrap the `RouteBuilder` if you need those.
- `AsyncBuilder` cannot detect when its `Load` closure has changed (Go function values are not comparable). To force a fresh invocation when inputs change, wrap it in `widgets.WithKey` with a key derived from those inputs.
- No InheritedWidget-style ambient dependency injection. Cross-tree state sharing goes through `gutter.Notifier` + `widgets.ObserverBuilder` — the app holds the `Notifier` and passes the pointer down explicitly.

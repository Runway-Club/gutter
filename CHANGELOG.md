# Changelog

## v0.5.0 — Full-stack: SSR, hydration, RPC

This release turns Gutter from a client-only WASM UI library into a full-stack
framework. The headline: server-side rendering with hydration closes the
cold-start gap with React (SSR first paint beats React at every size tier in the
`bench/` suite), and a typed RPC layer makes the client↔server boundary
type-safe with no codegen.

### Added

- **Server-side rendering.** `gutter.RenderToHTML(root)` walks the widget tree to
  HTML in pure Go (no `syscall/js`). `gutter.ServeSSR` / `gutter.SSRHandler`
  serve it.
- **Hydration.** `gutter.WithHydrate()` makes the WASM client adopt
  server-rendered DOM instead of rebuilding it — instant first paint, then
  interactive, no flash.
- **One-`main` entry: `gutter.Serve(gutter.Config{Root, RPC, Theme, …})`.** The
  same program compiles to both the WASM client and the host SSR server; `gutter
  run` serves it CSR, `gutter run --ssr` serves it server-rendered.
- **Typed RPC (`github.com/Runway-Club/gutter/rpc`).** `rpc.Handle(fn)` (server)
  and `rpc.Call[Req, Res](ctx, req)` (client), keyed by the request type — share
  Go structs across the boundary; changing a field is a compile error on both
  sides. No codegen, no string routes.
- **Dependency injection.** `gutter.Provider[T]` / `gutter.DependOn[T]` —
  InheritedWidget-style ambient values, correct under isolated `SetState` rebuilds.
- **Islands.** `gutter.MountInto` / `gutter.MountWhenVisible` — mount independent
  widget trees into an existing page; lazy-load WASM on viewport visibility.
- **Forms.** `widgets.Form` / `widgets.FormField` + composable `Validator`s
  (`Required`, `MinLength`, `MaxLength`, `Email`, `Pattern`, `Combine`).
- **Benchmark suite (`bench/`).** Render/reload/memory/CPU/compute and SSR-vs-React
  comparisons, with findings in `bench/ANALYSIS.md`.
- **Examples.** `examples/fullstack` (SSR + RPC, one `main.go`) and
  `examples/islands`.

### Changed

- **Accessibility.** `Heading` now renders semantic `<h1>`–`<h6>`; `Link` takes a
  real `Href`; `Image`/`IconButton` carry accessible names.
- **Router.** `Router.Query()` parses query strings, and route matching strips the
  query (`/user/42?tab=x` matches `/user/:id`).
- **CLI.** `gutter run --ssr` runs your `gutter.Serve` program as an SSR server;
  `gutter build deploy` defaults to TinyGo when installed (`--pure-go` opts out);
  `gutter new` scaffolds the unified `gutter.Serve` entry.

### Notes

- `RunApp` remains the supported low-level client entry; `Serve` is the
  batteries-included wrapper.
- The WASM bundle is large (Go ~3.4 MB / TinyGo ~1.2 MB raw); prefer TinyGo for
  production. SSR makes first paint fast regardless, but time-to-interactive still
  waits for the wasm to download and hydrate.

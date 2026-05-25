---
title: Server-side rendering & full-stack
nav_order: 8
---

# Server-side rendering & full-stack
{: .no_toc }

Build a complete web app — fast first paint, SEO, and a type-safe client↔server
boundary — from **one Go codebase and one `main`**.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## The one-`main` model

A Gutter app's client (WebAssembly, runs in the browser) and server (host
binary, runs SSR) are two different programs for two different platforms. You
do **not** write them separately. You write one `main` that calls
`gutter.Serve`, and the CLI compiles it twice:

```go
package main

import "github.com/Runway-Club/gutter"

func main() {
    gutter.Serve(gutter.Config{Root: Root})
}

func Root() gutter.Widget { /* your top widget */ }
```

`gutter.Serve` is build-tag aware:

| Build target | What `Serve` does |
| --- | --- |
| `GOOS=js GOARCH=wasm` (client) | `RunApp(Root(), WithHydrate())` — mounts the tree, or **hydrates** server-rendered HTML if present |
| host (server) | registers your RPC handlers, then serves SSR HTML at `/`, the RPC endpoint at `/rpc`, and the wasm assets for hydration |

This is exactly what `gutter new` scaffolds, so a fresh project already works in
both modes with zero changes.

`gutter.Config` fields (all optional except `Root`):

| Field | Purpose | Default |
| --- | --- | --- |
| `Root func() Widget` | builds the UI (client + SSR) — **required** | — |
| `RPC func()` | registers RPC handlers; runs on the server only | none |
| `Theme *themes.Theme` | active theme for client + SSR | `themes.Apple` |
| `Selector string` | client mount point | `#app` |
| `Addr string` | SSR listen address | `:8080` (env `GUTTER_ADDR` wins) |
| `Dist string` | SSR static-asset dir | `dist` (env `GUTTER_DIST` wins) |
| `Head string` | extra raw HTML in the SSR `<head>` | none |

---

## Step 1 — create and run (CSR)

```sh
gutter new myapp --module github.com/me/myapp
cd myapp
gutter run
```

Open <http://localhost:8080>. This is **client-side rendering (CSR)**: the
browser loads `app.wasm`, which builds the DOM. Simple, but the page is blank
until the wasm downloads and runs.

---

## Step 2 — turn on SSR

```sh
gutter run --ssr
```

Same code, one flag. Now the CLI:

1. builds the wasm into `dist/` (as before), then
2. runs your **same `main`** compiled for the host — `gutter.Serve`'s server
   renders `Root()` to HTML on every request and serves it.

The browser receives a fully-rendered page, so **content paints immediately**
(good FCP, crawlable for SEO) — then `app.wasm` loads in the background and
**hydrates** the existing DOM (wires up event handlers) without re-rendering or
flashing. Once hydrated, interactivity is identical to CSR.

> **How much faster?** On the benchmark in `bench/`, SSR first-contentful-paint
> beats both Gutter CSR and React at every size tier. See `bench/ANALYSIS.md` §8.

Nothing in your widget code changes between CSR and SSR. The same `Root()` runs
on the server (to produce HTML) and on the client (to hydrate).

---

## Step 3 — call the server with typed RPC

This is the payoff of one-language full-stack: call a server function from the
client with **shared Go types**, no codegen, no REST boilerplate, no stringly-
typed routes. Change a field and **both sides fail to compile**.

**a) Define the request/response types once**, in a package both sides import:

```go
// api/api.go
package api

type AddRequest struct{ A, B int }
type AddResponse struct{ Sum int }
```

**b) Register the handler in `Config.RPC`** (runs on the server only):

```go
// main.go
import (
    "context"
    "github.com/me/myapp/api"
    "github.com/Runway-Club/gutter"
    "github.com/Runway-Club/gutter/rpc"
)

func main() {
    gutter.Serve(gutter.Config{
        Root: Root,
        RPC: func() {
            rpc.Handle(func(_ context.Context, r api.AddRequest) (api.AddResponse, error) {
                return api.AddResponse{Sum: r.A + r.B}, nil
            })
        },
    })
}
```

**c) Call it from the client** — no URL, no fetch boilerplate. The route is
derived from the request type, so it always matches the handler:

```go
res, err := rpc.Call[api.AddRequest, api.AddResponse](ctx, api.AddRequest{A: 2, B: 40})
// res.Sum == 42
```

`rpc.Call` blocks on the network, so call it from a goroutine and `SetState`
the result back in:

```go
func (s *state) compute() {
    go func() {
        res, err := rpc.Call[api.AddRequest, api.AddResponse](
            context.Background(), api.AddRequest{A: 2, B: 40})
        s.SetState(func() {
            if err != nil { s.result = err.Error() } else { s.result = fmt.Sprint(res.Sum) }
        })
    }()
}
```

Run it with `gutter run --ssr` — the server serves both the page and `/rpc`.
The full working version is [`examples/fullstack`](examples.html).

> **Validation across the boundary.** Because validators are plain functions
> (see [Forms](#forms)), the exact same `widgets.Required`/`Email`/… checks can
> run client-side in a `Form` and server-side in the handler — one source of truth.

---

## Step 4 — deploy

```sh
gutter build deploy
```

Builds `dist/` (defaulting to **TinyGo** when it's installed, for a much smaller
bundle), writes a `Dockerfile` + `nginx.conf`, and builds the image. For an
SSR/RPC app you run the host server binary instead of static files; see
[CLI](cli.html). Pass `--pure-go` to force the standard toolchain.

---

## Project layout

A full-stack app is just:

```text
myapp/
├── go.mod
├── main.go        # func main() { gutter.Serve(gutter.Config{Root, RPC}) }
├── api/           # request/response structs shared by client + server
│   └── api.go
└── app or inline  # Root() and your widgets
```

No `main_wasm.go`, no `server/` directory, no build tags you write. The CLI owns
the two compilations.

---

## Islands — embed Gutter in an existing page

If you only want a few interactive widgets inside an otherwise static (or
non-Gutter) HTML page, use **islands** instead of owning the whole page:

```go
func main() {
    gutter.MountInto("#cart", CartWidget{}, gutter.WithHydrate())
    gutter.MountInto("#search", SearchWidget{}, gutter.WithHydrate())
    select {} // keep the runtime alive
}
```

`MountInto` is non-blocking, so you mount several independent trees and then
`select{}`. Pair it with a tiny page-side loader that fetches `app.wasm` only
when an island scrolls near the viewport — the static page then costs **zero
WASM** until interactivity is actually needed. `MountWhenVisible` defers an
individual island's mount the same way. Full example: [`examples/islands`](examples.html).

---

## Sharing values down the tree (DI)

Instead of threading a store, the RPC base URL, or feature flags through every
widget by hand, provide them once and read them anywhere:

```go
gutter.Provider[*Store]{Value: store, Child: /* ...app... */}

// deep inside the subtree, in any Build:
store, ok := gutter.DependOn[*Store](ctx)
```

Lookups are by exact type (define distinct named types for two values of the
same underlying type; a deeper `Provider[T]` shadows an outer one). It works
correctly even during isolated `SetState` rebuilds. For values that change
often, provide a `gutter.Notifier[T]` and read it with `widgets.ObserverBuilder`.

---

## Forms {#forms}

`widgets.Form` is a controlled, validated form — it owns field values + errors,
shows inline errors, and fires `OnSubmit` only when every field validates:

```go
widgets.Form{
    Fields: []widgets.FormField{
        {Name: "email", Label: "Email", Type: widgets.InputEmail,
         Validators: []widgets.Validator{
             widgets.Required("Email is required"),
             widgets.Email("Enter a valid email"),
         }},
    },
    Submit:   "Sign up",
    OnSubmit: func(v map[string]string) { /* rpc.Call, navigate, … */ },
}
```

Validators (`Required`, `MinLength`, `MaxLength`, `Email`, `Pattern`, `Combine`)
are plain `func(string) string` — compose them, and reuse the same ones in your
RPC handlers.

---

## Gotchas

- **Bundle size.** The wasm is large (Go ~3.4 MB / TinyGo ~1.2 MB raw; gzip ~950
  KB / ~410 KB) versus a typical JS bundle. SSR makes *first paint* fast
  regardless, but *time-to-interactive* still waits for the wasm to download and
  hydrate — so prefer **TinyGo for production** (`gutter build deploy` defaults
  to it).
- **Server-only code in the client bundle.** Handlers referenced in `Config.RPC`
  are linked into the wasm client too (dead, but present). Keep heavy server-only
  logic (DB access, secrets) in files tagged `//go:build !js || !wasm` so it
  never reaches the browser, and have the handler call into it.
- **Security.** RPC handlers are public HTTP endpoints. Validate and authorize
  inside the handler; never trust the client.

---

## Next steps

- **[Examples](examples.html)** — `fullstack` (SSR + RPC) and `islands` annotated.
- **[State Management](state-management.html)** — `SetState`, lifecycle, keyed lists.
- **[CLI](cli.html)** — every command and flag, including `run --ssr`.
- **[Architecture](architecture.html)** — how SSR, hydration, and the reconciler work.

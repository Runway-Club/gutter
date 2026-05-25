---
title: Getting Started
nav_order: 2
---

# Getting Started
{: .no_toc }

This guide takes you from zero to a running Gutter app in your browser.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Prerequisites

- **Go 1.21 or newer.** Gutter targets `GOOS=js GOARCH=wasm`; you don't need any extra toolchain — the standard Go compiler can already produce WebAssembly. Go 1.24+ ships `wasm_exec.js` under `$GOROOT/lib/wasm/`; older Go ships it under `$GOROOT/misc/wasm/`. Gutter's CLI handles both.
- **A modern browser.** Anything that supports WebAssembly and `fetch` — Chrome, Safari, Firefox, Edge.

You do **not** need Node, npm, Webpack, Vite, or any JavaScript build tool. The only JavaScript Gutter ships is the tiny `wasm_exec.js` glue file provided by the Go distribution.

---

## Install the CLI

```sh
go install github.com/Runway-Club/gutter/cmd/gutter@latest
```

This installs the `gutter` binary into `$GOBIN` (or `$GOPATH/bin`). Make sure that directory is on your `PATH`:

```sh
gutter --version
# gutter version 0.5.0
```

If you are working from a local clone of the repo instead, build the CLI from source:

```sh
git clone https://github.com/Runway-Club/gutter
cd gutter
go build -o bin/gutter ./cmd/gutter
./bin/gutter --version
```

---

## Scaffold your first project

```sh
gutter new myapp
cd myapp
```

`gutter new` walks you through a short interactive prompt (project name + Go module path) if you skip the arguments. It writes four files:

```text
myapp/
├── .gitignore     # Ignores ./dist/ and any stray top-level build artifacts
├── go.mod         # Go module declaration (pinned to the latest Gutter release)
├── index.html     # HTML host page (loads Lexend + wasm_exec.js + your app)
└── main.go        # Your app — a Root() widget builder + gutter.Serve
```

After writing them, the CLI runs `go get github.com/Runway-Club/gutter@latest` inside the new project, so `go.mod` is pinned to the current published version (no manual `go mod tidy` needed). The `main.go` it writes is a complete, runnable "Hello, Gutter!" — a `Scaffold` with an `AppBar`, a centered `Card`, a `Heading`, a `Body`, and a primary `Button` — wired through `gutter.Serve`, so the **same project runs client-side (`gutter run`) or server-side (`gutter run --ssr`)** with no changes.

> **Working from a local checkout?** `gutter new` does not emit a `replace` directive. If your `go.mod` should point at your local Gutter clone instead of the published module, add the directive yourself:
>
> ```go
> replace github.com/Runway-Club/gutter => ../path/to/gutter
> ```

---

## Run it

```sh
gutter run
# or, with live reload on .go / .html / .css change:
gutter run dev
```

`gutter run` compiles your app and bundles `app.wasm`, your `index.html`, `wasm_exec.js`, and anything in `./public/` into `./dist/`, registers the `application/wasm` MIME type, and serves `./dist/` at <http://localhost:8080>. Your project root stays clean — all build artifacts live under `./dist/`.

`gutter run dev` does the same, but also watches `.go`, `.html`, and `.css` files, rebuilds on save, and injects a small reload-poller script into `index.html` so your browser refreshes automatically when the build counter ticks.

Open <http://localhost:8080>. You should see the scaffolded landing page rendered with the Apple theme.

---

## The whole app, top to bottom

Open `main.go` and you'll see the canonical Gutter shape:

```go
package main

import (
    "github.com/Runway-Club/gutter"
    "github.com/Runway-Club/gutter/themes"
    "github.com/Runway-Club/gutter/widgets"
)

// Root builds the app's UI. gutter.Serve calls it on the client to mount/hydrate
// and on the server to render HTML (gutter run --ssr).
func Root() gutter.Widget {
    return widgets.Scaffold{
        Title:  "myapp",
        Theme:  themes.Apple,
        AppBar: widgets.AppBar{Title: "myapp"},
        Body: widgets.Surface{
            Variant: widgets.SurfaceAlt,
            Child:   widgets.Center{Child: /* ... */},
        },
    }
}

func main() { gutter.Serve(gutter.Config{Root: Root}) }
```

The three things worth noticing:

1. **`Root` returns a `Scaffold`.** `Scaffold` is a `StatelessWidget` — it owns the app-wide theme, sets the document title, and lays out the app bar + body + optional footer. You'll almost always start here.
2. **`gutter.Serve(gutter.Config{Root: Root})` is the one entry point.** It works for both modes: `gutter run` mounts the tree client-side; `gutter run --ssr` builds the wasm and runs this same program as a server that renders `Root()` to HTML, then the client hydrates it. No build tags, no second `main`.
3. **It's all Go.** No router config, no JSX, no CSS files.

> Prefer the lowest-level entry? `gutter.RunApp(Root())` still works for a pure client-side app — `Serve` is just the batteries-included wrapper that also knows how to be an SSR server.

That's the whole app.

---

## Add interactivity

Stateless widgets can't change between rebuilds because they have no place to put state. The moment you need a counter, a toggle, a form value, switch to a **StatefulWidget**:

```go
package main

import (
    "fmt"

    "github.com/Runway-Club/gutter"
    "github.com/Runway-Club/gutter/widgets"
)

type CounterApp struct{}

func (CounterApp) CreateState() gutter.State { return &counterState{} }

type counterState struct {
    gutter.StateObject
    count int
}

func (s *counterState) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Center{
        Child: widgets.Column{
            CrossAxisAlign: widgets.CrossAxisCenter,
            Spacing:        16,
            Children: []gutter.Widget{
                widgets.Heading{
                    Level: widgets.H2,
                    Text:  fmt.Sprintf("Count: %d", s.count),
                },
                widgets.Button{
                    Variant:   widgets.ButtonPrimary,
                    Label:     "Increment",
                    OnPressed: func() { s.SetState(func() { s.count++ }) },
                },
            },
        },
    }
}

func Root() gutter.Widget { return CounterApp{} }

func main() { gutter.Serve(gutter.Config{Root: Root}) }
```

`StateObject` is the SetState mixin — embed it by value, return a **pointer** to your state struct from `CreateState`. The framework binds the element handle automatically, so `s.SetState(fn)` runs `fn` and then rebuilds **only the subtree this State owns**. Sibling widgets, focused inputs, and even other StatefulWidgets are untouched.

See [State Management](state-management.html) for the full lifecycle (`InitState`, `Dispose`), keyed lists, and how the reconciler decides what to reuse.

---

## Server-side rendering — one flag

Because your `main` uses `gutter.Serve`, you can render on the server with no code changes:

```sh
gutter run --ssr
```

The CLI builds the wasm, then runs your same program as a server that renders `Root()` to HTML per request. The browser **paints content immediately** (better first paint, crawlable for SEO), then `app.wasm` loads and **hydrates** the page into a live, interactive app. You can also call server functions from the client with type-safe RPC — sharing Go structs across the boundary, no codegen.

This — SSR, hydration, typed RPC, islands — is covered end-to-end in **[Server-side rendering & full-stack](fullstack.html)**.

---

## Build for production

```sh
gutter build
```

Writes a self-contained bundle to `./dist/`:

```text
dist/
├── app.wasm        # your compiled app
├── index.html      # copied from current directory
├── wasm_exec.js    # from $GOROOT
└── …               # plus anything in ./public/, if present
```

Drop `dist/` behind any static file server — nginx, Caddy, Cloudflare Pages, GitHub Pages, an S3 bucket, whatever — as long as it serves `.wasm` files with `Content-Type: application/wasm`.

For a Docker image with nginx pre-configured, run:

```sh
gutter build deploy
```

This produces `./dist/`, generates a `Dockerfile`, `nginx.conf`, and `.dockerignore` (only if missing), and runs `docker build`. `build deploy` defaults to **TinyGo** when it's installed (a much smaller production bundle); pass `--pure-go` to opt out. See [CLI](cli.html) for details.

---

## Next steps

- **[Server-side rendering & full-stack](fullstack.html)** — SSR, hydration, typed client↔server RPC, islands.
- **[Architecture](architecture.html)** — the widget model, the persistent element tree, how reconciliation works.
- **[State Management](state-management.html)** — `StateObject`, lifecycle hooks, keyed lists.
- **[Themes](themes.html)** — built-in presets, switching themes, the token tables.
- **[Widgets](widgets/)** — every widget, what it does, when to reach for it.

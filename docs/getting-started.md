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
# gutter version 0.2.0
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

`gutter new` walks you through a short interactive prompt (project name + Go module path) if you skip the arguments. It writes three files:

```text
myapp/
├── go.mod         # Go module declaration
├── index.html     # HTML host page (loads Lexend + wasm_exec.js + your app)
└── main.go        # Your app — Build a widget tree, call RunApp
```

The `main.go` it writes is a complete, runnable "Hello, Gutter!" — a `Scaffold` with an `AppBar`, a centered `Card`, a `Heading`, a `Body`, and a primary `Button`.

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

`gutter run` compiles `./main.go` to `./app.wasm`, copies `wasm_exec.js` next to it, registers the `application/wasm` MIME type, and serves the current directory at <http://localhost:8080>.

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

type App struct{}

func (App) Build(ctx *gutter.BuildContext) gutter.Widget {
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

func main() { gutter.RunApp(App{}) }
```

The three things worth noticing:

1. **`App` is a `StatelessWidget`.** It satisfies the interface by having a single `Build(ctx) Widget` method. The framework calls it once at mount time and again whenever an ancestor rebuilds.
2. **`Scaffold` is the root.** It owns the app-wide theme, sets the document title, and lays out the app bar + body + optional footer. You'll almost always start here.
3. **`gutter.RunApp(App{})` mounts the tree.** It finds `#app` in `index.html`, hands the BuildContext (with the active theme) down to your widget, and then blocks the goroutine forever so JS callbacks keep firing.

That's the whole app. No router, no JSX, no CSS files.

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

func main() { gutter.RunApp(CounterApp{}) }
```

`StateObject` is the SetState mixin — embed it by value, return a **pointer** to your state struct from `CreateState`. The framework binds the element handle automatically, so `s.SetState(fn)` runs `fn` and then rebuilds **only the subtree this State owns**. Sibling widgets, focused inputs, and even other StatefulWidgets are untouched.

See [State Management](state-management.html) for the full lifecycle (`InitState`, `Dispose`), keyed lists, and how the reconciler decides what to reuse.

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

This produces `./dist/`, generates a `Dockerfile`, `nginx.conf`, and `.dockerignore` (only if missing), and runs `docker build`. See [CLI](cli.html) for details.

---

## Next steps

- **[Architecture](architecture.html)** — the widget model, the persistent element tree, how reconciliation works.
- **[State Management](state-management.html)** — `StateObject`, lifecycle hooks, keyed lists.
- **[Themes](themes.html)** — built-in presets, switching themes, the token tables.
- **[Widgets](widgets/)** — every widget, what it does, when to reach for it.

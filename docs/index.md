---
title: Home
layout: home
nav_order: 1
description: "Gutter is a declarative web framework for Go that compiles to WebAssembly and drives the browser DOM directly."
permalink: /
---

<p align="center">
  <img src="{{ '/assets/gutter_icon.png' | relative_url }}" alt="Gutter" width="120" />
</p>

# Gutter
{: .fs-9 }

Gutter is a **declarative web framework for Go**. Compose your interface from widgets — Flutter-inspired — and the runtime compiles to WebAssembly and drives the browser DOM directly. No CSS in your app code, no JavaScript toolchain.
{: .fs-6 .fw-300 }

[Get started now](getting-started.html){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View on GitHub](https://github.com/Runway-Club/gutter){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## What is Gutter?

Gutter is a declarative web framework for Go. It brings the **declarative widget model** popularized by Flutter and React to Go on the web: you describe what the UI should look like as a tree of widgets, and the runtime takes care of materializing it into a persistent DOM tree, diffing updates, and wiring up events — all from a single Go binary compiled to WebAssembly.

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
        Title: "Hello, Gutter!",
        Theme: themes.Apple,
        AppBar: widgets.AppBar{Title: "Hello"},
        Body: widgets.Center{
            Child: widgets.Card{
                Variant: widgets.CardFeature,
                Child: widgets.Column{
                    CrossAxisAlign: widgets.CrossAxisCenter,
                    Spacing:        16,
                    Children: []gutter.Widget{
                        widgets.Heading{Level: widgets.H2, Text: "Hello, Gutter!"},
                        widgets.Body{Text: "Pick a theme and ship — no CSS needed."},
                        widgets.Button{Variant: widgets.ButtonPrimary, Label: "Get started"},
                    },
                },
            },
        },
    }
}

func main() { gutter.RunApp(App{}) }
```

---

## Why Gutter?

- **One language, top to bottom.** Write your server in Go and your UI in Go. Share types and validation logic across the boundary instead of duplicating them in TypeScript.
- **No `node_modules`.** No npm, no webpack, no bundler config, no `node_modules` to audit or keep patched. One `go.mod` and the `gutter` CLI are the whole toolchain.
- **Fast, small bundles.** Opt into TinyGo with `--tinygo` for **4–8× smaller** WebAssembly — a counter app drops from ~2.8 MB to ~340 KB — so the browser downloads and instantiates less before first paint.
- **Heavy work stays on the client.** Offload CPU-bound tasks to a Web Worker with `Worker` + `gutter.NewWorkerTask` — parse files, crunch data, run long computations — without blocking the UI thread, all in Go.
- **Declarative widgets, persistent element tree.** Gutter mirrors Flutter and React: widgets are immutable descriptions, the runtime keeps a persistent element tree, and `SetState` rebuilds only the subtree that changed.
- **Themes do the styling.** Every themed widget reads from the active theme. Pick `themes.Apple` or `themes.Meta`, and a `Button{Variant: ButtonPrimary}` becomes an Action Blue pill or a black pill — without changing a line of app code.
- **No CSS in app code.** Themed widgets resolve colors, typography, shape, and spacing from theme tokens. Layout primitives (`Column`, `Row`, `Center`, `Padding`) don't need CSS either.
- **A real CLI.** `gutter new` scaffolds a project, `gutter run dev` gives you live reload on save, `gutter build deploy` produces a `dist/` plus a Dockerfile.

---

## Where next?

| If you want to…                       | Read                                                           |
| ------------------------------------- | -------------------------------------------------------------- |
| Install Gutter and run your first app | [Getting Started](getting-started.html)                        |
| Understand how it works               | [Architecture](architecture.html)                              |
| Build interactive screens             | [State Management](state-management.html)                      |
| Switch or extend a theme              | [Themes](themes.html)                                          |
| Ship static files with your app       | [Assets](assets.html)                                          |
| Browse every widget                   | [Widgets](widgets/)                                            |
| Use vendor-specific widgets           | [Community packages](community.html)                           |
| Use the developer CLI                 | [CLI](cli.html)                                                |
| Read annotated examples               | [Examples](examples.html)                                      |

---

## Status

Early prototype with a fast-growing widget catalog. Production-shaped API, single-pass reconciler, two real design systems (Apple, Meta) shipped as presets. A `community/` tier for vendor-specific reusable widgets (Google sign-in, …) lives next to the core catalog.

**Currently shipped**:

- App shell: `Scaffold` (with `StickyAppBar`), `AppBar`, `Heading`, `Body`, `Caption`, `Link`, `Button`, `IconButton`, `Card`, `Surface`, `Badge`, `Image`, `Icon` (Material Symbols), `File` (file picker that reads bytes via FileReader).
- Inputs: `Input` (13 HTML types — text, password, email, number, tel, url, search, date, time, datetime-local, month, week, color), `TextArea`, `Checkbox`, `Switch`, `Slider`, `Select[T]`, `RadioGroup[T]`. All controlled (declarative `Value`/`Checked`/`Selected` field is source of truth).
- Layout: `Column`, `Row`, `Center`, `Padding`, `SizedBox`, `Container`, `Styled`, `Transform`, `List`, `ListBuilder` (virtualized 10k+ row list with DOM recycling).
- Overlays: `Popup`, `Drawer`, `BottomSheet`.
- Reactive: `Notifier[T]`/`Listenable[T]`, `ObserverBuilder[T]`, `AsyncBuilder[T]`, `AnimationController` + `AnimatedBuilder`, `Router` + `RouterView`.
- Imperative: `Canvas` (typed 2D painter), `GestureDetector`, `Worker` (Web Worker with inline Go task).
- Assets: `gutter.AssetURL`, configurable base, CLI copies `./assets/` to `dist/assets/`.

**Known gaps**: no SSR, no tests, no devtools, `SetState` is synchronous and unbatched, `ListBuilder` requires fixed row heights, theme overrides per subtree need explicit `Scaffold` nesting (no InheritedWidget yet).

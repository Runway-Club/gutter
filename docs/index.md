---
title: Home
layout: home
nav_order: 1
description: "Gutter is a Flutter-inspired declarative UI library for Go that compiles to WebAssembly and drives the browser DOM directly."
permalink: /
---

<p align="center">
  <img src="{{ '/assets/gutter_icon.png' | relative_url }}" alt="Gutter" width="120" />
</p>

# Gutter
{: .fs-9 }

A Flutter-inspired declarative UI library for Go. Compose your interface from widgets; the runtime compiles to WebAssembly and drives the browser DOM directly — no CSS in your app code.
{: .fs-6 .fw-300 }

[Get started now](getting-started.html){: .btn .btn-primary .fs-5 .mb-4 .mb-md-0 .mr-2 }
[View on GitHub](https://github.com/Runway-Club/gutter){: .btn .fs-5 .mb-4 .mb-md-0 }

---

## What is Gutter?

Gutter brings the **declarative widget model** popularized by Flutter and React to Go on the web. You describe what the UI should look like as a tree of widgets, and the runtime takes care of materializing it into a persistent DOM tree, diffing updates, and wiring up events — all from a single Go binary compiled to WebAssembly.

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
| Browse every widget                   | [Widgets](widgets/)                                            |
| Use the developer CLI                 | [CLI](cli.html)                                                |
| Read annotated examples               | [Examples](examples.html)                                      |

---

## Status

Early prototype. Production-shaped API, single-pass reconciler, two real design systems (Apple, Meta) shipped as presets. Known gaps: no router, no animation framework, no image widget, no SSR — `SetState` is synchronous and unbatched.

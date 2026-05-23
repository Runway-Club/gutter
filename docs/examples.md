---
title: Examples
nav_order: 7
---

# Examples
{: .no_toc }

Two runnable example apps ship in [`examples/`](https://github.com/Runway-Club/gutter/tree/main/examples). They're not just demos — they're the smallest end-to-end illustration of every concept in the framework.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Running an example

Both examples use a `replace` directive in their own `go.mod` so they build against the working copy of the repository:

```sh
git clone https://github.com/Runway-Club/gutter
cd gutter/examples/counter
go run ../../cmd/gutter run
# open http://localhost:8080
```

Swap `counter` for `showcase` to run the other one.

---

## `examples/counter` — a stateful counter

[Source](https://github.com/Runway-Club/gutter/tree/main/examples/counter). The minimal interactive app, ~70 lines.

```go
type CounterApp struct{}

func (CounterApp) CreateState() gutter.State { return &counterState{} }

type counterState struct {
    gutter.StateObject
    count int
}

func (s *counterState) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Scaffold{
        Title: "Gutter Counter",
        Theme: themes.Meta,
        AppBar: widgets.AppBar{
            TitleWidget: widgets.Text{Data: "Counter", Style: &widgets.TextStyle{FontSize: "20px"}},
            Actions: []gutter.Widget{
                widgets.Button{
                    Variant:   widgets.ButtonGhost,
                    Label:     "Reset",
                    OnPressed: func() { s.SetState(func() { s.count = 0 }) },
                },
            },
        },
        Body: widgets.Surface{
            Variant: widgets.SurfaceAlt,
            Child: widgets.Center{
                Child: widgets.Card{
                    Variant: widgets.CardFeature,
                    Child: widgets.Column{
                        CrossAxisAlign: widgets.CrossAxisCenter,
                        Spacing:        16,
                        Children: []gutter.Widget{
                            widgets.Heading{Level: widgets.H2, Text: fmt.Sprintf("Count: %d", s.count)},
                            widgets.Body{Text: "Tap the buttons. No CSS in this file.", Small: true},
                            widgets.Row{
                                Spacing: 8,
                                Children: []gutter.Widget{
                                    widgets.Button{Variant: widgets.ButtonPrimary, Label: "−",
                                        OnPressed: func() { s.SetState(func() { s.count-- }) }},
                                    widgets.Button{Variant: widgets.ButtonPrimary, Label: "+",
                                        OnPressed: func() { s.SetState(func() { s.count++ }) }},
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}

func main() { gutter.RunApp(CounterApp{}) }
```

### What to notice

- **`CounterApp` is a `StatefulWidget`** — it has a `CreateState`, not a `Build`. The state lives on `counterState`, which embeds `gutter.StateObject` (the SetState mixin) by value.
- **`Build` is on `*counterState`** — pointer receiver. Mutating `s.count` on a value receiver wouldn't stick across calls.
- **`s.SetState(func() { s.count++ })`** mutates state and rebuilds just this state's subtree. The AppBar isn't rebuilt; the Surface isn't rebuilt; only the Column under the Card is.
- **`Scaffold.Theme: themes.Meta`** — Scaffold drives the theme. No `gutter.WithTheme` is needed at `RunApp`.
- **`fmt.Sprintf("Count: %d", s.count)`** — there's no template language. You're in Go, use Go.

---

## `examples/showcase` — one widget tree, two themes

[Source](https://github.com/Runway-Club/gutter/tree/main/examples/showcase). The same widget tree, rendered under whichever theme you pick at build time:

```sh
cd examples/showcase

# Default — Apple:
go run ../../cmd/gutter run

# Or, compile with Meta:
GOOS=js GOARCH=wasm go build -ldflags "-X 'main.themeName=meta'" -o app.wasm .
gutter run
```

The whole point of the showcase is to make the theme tax visible: identical Go code, completely different visual identity.

```go
var themeName = "apple"  // overridden via -ldflags

func main() {
    theme := themes.Apple
    if themeName == "meta" {
        theme = themes.Meta
    }
    gutter.RunApp(Showcase{}, gutter.WithTheme(theme))
}
```

### What to notice

- **A marketing-page layout from stacked Surfaces.** Hero band (`SurfaceDark`) → feature row (`SurfaceCanvas`) → promo card on alt surface (`SurfaceAlt` + `CardPromo`) → status row (`SurfaceCanvas` + badges).
- **Colors come from the theme on dark surfaces.** Inside `SurfaceDark` / `CardPromo`, the headings and body text pass `Color: ctx.Theme.Colors.OnDark` so they're legible.
- **A factory function** (`featureCard`) builds three near-identical cards from a `title` + `body`. There's no JSX-like template — just a Go function returning a widget.
- **All four badge variants** in a single Row, showing the semantic palette: neutral, success, warning, critical.

The pattern of "outer `Surface{Padding: "0"}` wrapping a `Column` of full-bleed `Surface`s" is worth memorizing — it's how every marketing page in this style is composed.

---

## Beyond the examples

The examples cover the framework's surface area. From here, the natural next steps are:

- **Write a custom widget.** A `StatelessWidget` that wraps a `Card` + `Column` to give your feature tile a name (`FeatureCard{Title, Body}`).
- **Build a small form.** A `StatefulWidget` with `Input` fields, `OnChanged → SetState`, a submit `Button`, and an `Error` boolean per field.
- **Write a list with reorderable items.** Use `widgets.WithKey{Key: item.ID, Child: …}` so the reconciler preserves each item's State across reorders.
- **Define your own theme.** Copy `themes/apple.go`, change the values, give the variable a new name, pass it to `Scaffold.Theme`.

If you build something interesting, the maintainers want to hear about it.

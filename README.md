# Gutter

A Flutter-inspired declarative UI library for Go. Compose your interface from
widgets; the runtime compiles to WebAssembly and drives the browser DOM
directly.

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
        Title: "Hello",
        Theme: themes.Apple,
        AppBar: widgets.AppBar{
            Title: "Hello",
            Actions: []gutter.Widget{
                widgets.Button{Variant: widgets.ButtonGhost, Label: "About"},
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
                            widgets.Heading{Level: widgets.H2, Text: "Hello, Gutter!"},
                            widgets.Body{Text: "Pick a theme and ship — no CSS needed."},
                            widgets.Button{Variant: widgets.ButtonPrimary, Label: "Get started"},
                        },
                    },
                },
            },
        },
    }
}

func main() {
    gutter.RunApp(App{})
}
```

The widget catalog is theme-aware by default — pick a variant, the theme picks the values. No CSS in app code.

| Package | What's in it |
| --- | --- |
| `github.com/Runway-Club/gutter` | Framework core: `Widget`, `Host`, `State`, `BuildContext`, runtime, `RunApp`, options. |
| `github.com/Runway-Club/gutter/themes` | Theme data — `Theme`, `Colors`, `Typography`, ...; ready-made `themes.Apple`, `themes.Meta`, `themes.Neutral`. |
| `github.com/Runway-Club/gutter/widgets` | The single widget catalog. Themed widgets (`Heading`, `Body`, `Button`, `Card`, `Surface`, `Input`, `Badge`, `Link`) read the active theme; layout primitives (`Column`, `Row`, `Center`, `Padding`, `SizedBox`) and escape-hatches (`Styled`, `Text`, `Container`) carry no theme dependency. |

## Installing the CLI

```sh
go install github.com/Runway-Club/gutter/cmd/gutter@latest
```

If you are working from a local clone, build the CLI from the repo:

```sh
go build -o bin/gutter ./cmd/gutter
```

## Quick start

```sh
gutter new myapp
cd myapp
gutter run
```

Then open <http://localhost:8080>.

## CLI

| Command          | What it does                                                       |
| ---------------- | ------------------------------------------------------------------ |
| `gutter new <n>` | Scaffold a project (`main.go`, `index.html`, `go.mod`)             |
| `gutter run`     | Build the current dir as WASM, copy `wasm_exec.js`, serve on :8080 |
| `gutter build`   | Build a production-ready bundle into `./dist`                      |

`wasm_exec.js` is copied from `$GOROOT/lib/wasm/` (Go 1.24+) or
`$GOROOT/misc/wasm/` (older).

## Themes

Two production design systems ship in the box, extracted from the specs in
`theme_specs/`. Apple is the default — `gutter.RunApp(MyApp{})` with no
options uses it.

| Theme | Vibe |
| --- | --- |
| `themes.Apple` (default) | Photography-first museum gallery — single Action Blue accent, parchment + dark alternating tiles, pill primary CTAs. |
| `themes.Meta` | Hardware merchandiser — black-pill marketing primary + cobalt commerce primary, 32px rounded photographic cards. |
| `themes.Neutral` | Lexend-only neutral fallback for tests / brand-agnostic apps. |

Switch themes at app startup:

```go
gutter.RunApp(MyApp{}, gutter.WithTheme(themes.Meta))
```

The same widget tree renders differently under each theme: button variants,
heading sizes, card geometry, all driven from the token tables. See
`examples/showcase/main.go` for a side-by-side demonstration (build with
`-ldflags "-X main.themeName=meta"` to flip).

**Font.** Every built-in theme leads its font stack with [Lexend](https://www.lexend.com/),
loaded from Google Fonts in the scaffolded `index.html`. SF Pro / Optimistic
VF remain as platform fallbacks behind it. If you scaffold via `gutter new`,
the Lexend `<link>` is already in `index.html`; if you write your own HTML,
add it yourself.

## Widgets

All widgets live in `github.com/Runway-Club/gutter/widgets`. Themed widgets read
the active theme from `BuildContext`; layout primitives carry no theme
dependency.

| Widget      | Theme-aware | Variants / notes                                              |
| ----------- | :---------: | ------------------------------------------------------------- |
| `Scaffold`  |     yes     | The app shell — `Title`, `Theme`, `AppBar`, `Body`, `Footer`. |
| `AppBar`    |     yes     | Top nav strip — `Title`, `Leading`, `Actions[]`               |
| `Heading`   |     yes     | `H1`–`H6` — display and heading typography                    |
| `Body`      |     yes     | `Bold`, `Small`                                               |
| `Caption`   |     yes     | Shorthand for `Body{Small:true}`                              |
| `Link`      |     yes     | Themed inline anchor                                          |
| `Button`    |     yes     | `Primary`, `Secondary`, `Ghost`, `Accent`, `OnDark`           |
| `Card`      |     yes     | `Feature`, `Promo`, `Plain`                                   |
| `Surface`   |     yes     | `Canvas`, `Alt`, `Dark` — full-bleed regions                  |
| `Input`     |     yes     | Themed text field with `Error`                                |
| `Badge`     |     yes     | `Neutral`, `Success`, `Warning`, `Critical`                   |
| `Column`    |      no     | Vertical flex with `Spacing`                                  |
| `Row`       |      no     | Horizontal flex with `Spacing`                                |
| `Center`    |      no     | Centers a single child                                        |
| `Padding`   |      no     | Wraps a child with `EdgeInsets`                               |
| `SizedBox`  |      no     | Fixed width/height                                            |
| `Container` |      no     | Low-level styled `<div>` (raw colors/borders/radii)           |
| `Text`      |      no     | Raw `<span>` with explicit `TextStyle`                        |
| `Styled`    |      no     | Escape hatch — any tag, arbitrary attrs/style/events          |
| `WithKey`   |      no     | Wraps a child with a reconciliation key                       |

### Scaffold

`widgets.Scaffold` is the recommended root of every app. It picks the theme,
sets the document title, lays out the app bar above the body, and (when
present) a footer below. Pass the theme on Scaffold and skip
`gutter.WithTheme` at `RunApp` — Scaffold wins.

```go
widgets.Scaffold{
    Title:  "My App",        // → document.title
    Theme:  themes.Meta,     // → ctx.Theme for the whole tree
    AppBar: widgets.AppBar{Title: "My App"},
    Body:   widgets.Center{Child: ...},
    Footer: nil,              // optional
}
```

`AppBar` reads the theme's NavBar style: Apple ships a 44px black band with
12px white nav labels; Meta ships a 64px white bar with a hairline-soft
bottom border and a 14px/700 button-style label.

### Stateful widgets

Embed `gutter.StateObject` to gain `SetState`. The runtime maintains a
persistent Element tree (Flutter-style), so `SetState` rebuilds **only the
subtree owned by that State** — DOM nodes are diffed and updated in place,
event listeners are re-registered, and unrelated parts of the tree (including
focused inputs) are not touched.

```go
type CounterApp struct{}

func (CounterApp) CreateState() gutter.State { return &counterState{} }

type counterState struct {
    gutter.StateObject
    count int
}

func (s *counterState) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Button{
        Label:     fmt.Sprintf("Count: %d", s.count),
        OnPressed: func() { s.SetState(func() { s.count++ }) },
    }
}
```

`CreateState` must return a **pointer** to your state struct so the embedded
`StateObject` can be mutated by the framework.

## Examples

Run from a checkout of this repo:

```sh
cd examples/counter
go run ../../cmd/gutter run

# or the cross-theme showcase
cd examples/showcase
go run ../../cmd/gutter run
```

## Keyed reconciliation

For lists where the order can change, give each child a key. Either implement
`gutter.Keyed` on your widget, or wrap it with `widgets.WithKey`:

```go
widgets.Column{
    Children: []gutter.Widget{
        widgets.WithKey{Key: todo.ID, Child: TodoItem{Todo: todo}},
        // ...
    },
}
```

Without keys, the reconciler matches unkeyed siblings of the same Go type
positionally — fine for stable lists but wrong for reorder/insert/delete in
the middle.

## Status

Early prototype. Known gaps:

- `SetState` is synchronous and unbatched.
- No `Image`, no routing, no theming, no animation.
- No tests; WASM tests need a browser harness.

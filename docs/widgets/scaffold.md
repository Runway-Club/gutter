---
title: Scaffold
parent: Widgets
nav_order: 1
---

# `Scaffold`
{: .no_toc }

The recommended root of every Gutter app. It ties together the four big pieces of a real app: title, theme, app bar, body, and optional footer.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Scaffold struct {
    Title  string
    Theme  *themes.Theme
    AppBar gutter.Widget
    Body   gutter.Widget
    Footer gutter.Widget
}
```

| Field    | What it does                                                                 |
| -------- | ---------------------------------------------------------------------------- |
| `Title`  | Pushed to `document.title` (via `gutter.SetTitle`).                          |
| `Theme`  | Mutates `ctx.Theme` for the whole subtree — wins over `gutter.WithTheme`.    |
| `AppBar` | The top navigation strip. Use [`widgets.AppBar`](appbar.html).               |
| `Body`   | Your main content. Gets `flex: 1` — takes the remaining vertical space.     |
| `Footer` | Optional bottom strip. Renders below `Body`.                                 |

Background and ink colors are pulled from the active theme's `Colors.Canvas` and `Colors.Ink`. Layout is `display: flex; flex-direction: column; min-height: 100%`, so a `Center` inside `Body` fills the viewport minus the chrome.

---

## Basic usage

```go
type App struct{}

func (App) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Scaffold{
        Title:  "Hello",
        Theme:  themes.Apple,
        AppBar: widgets.AppBar{Title: "Hello"},
        Body: widgets.Center{
            Child: widgets.Heading{Level: widgets.H2, Text: "Hi!"},
        },
    }
}

func main() { gutter.RunApp(App{}) }
```

---

## With a footer

```go
widgets.Scaffold{
    Title:  "My App",
    Theme:  themes.Meta,
    AppBar: widgets.AppBar{Title: "My App"},
    Body:   widgets.Surface{Variant: widgets.SurfaceAlt, Child: /* … */},
    Footer: widgets.Surface{
        Variant: widgets.SurfaceDark,
        Padding: "16px 24px",
        Child: widgets.Caption{
            Text:  "© 2025 My Company",
            Color: themes.Meta.Colors.OnDark,
        },
    },
}
```

---

## With a full-bleed body

A `Surface` directly inside `Body` becomes a hero band — it takes full width and the body's full available height:

```go
widgets.Scaffold{
    AppBar: widgets.AppBar{Title: "Landing"},
    Body: widgets.Surface{
        Variant: widgets.SurfaceDark,
        Child: widgets.Center{
            Child: widgets.Heading{
                Level: widgets.H1,
                Text:  "One catalog. Two design systems.",
                Color: ctx.Theme.Colors.OnDark,
            },
        },
    },
}
```

---

## Theme precedence

`Scaffold.Theme` wins over `gutter.WithTheme` because it runs **during the Build**, after `RunApp` has already populated `ctx.Theme`:

```go
gutter.RunApp(App{}, gutter.WithTheme(themes.Meta))
// where App.Build returns Scaffold{Theme: themes.Apple, …}
// → renders with Apple, not Meta.
```

Recommendation: set the theme on `Scaffold` and skip `WithTheme` at `RunApp`. Keeps the shell configuration in one place.

If `Scaffold.Theme` is nil, the existing `ctx.Theme` is kept untouched (so `WithTheme` still works, and `themes.Apple` is the framework fallback).

---

## Notes

- `Scaffold` is a `StatelessWidget` — there's no `Scaffold` state to manage.
- `Title` defaults to an empty string. If empty, `document.title` is left alone.
- All fields except `Title`/`Theme` are optional. A `Scaffold{}` with nothing in it renders just the themed canvas color.

---

## See also

- [`AppBar`](appbar.html) — the top nav strip you'll usually pair with Scaffold.
- [`Surface`](surface.html) — the natural body for hero / banded layouts.
- [Themes](../themes.html) — the data structure Scaffold's `Theme` field accepts.

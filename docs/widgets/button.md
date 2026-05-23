---
title: Button
parent: Widgets
nav_order: 7
---

# `Button`
{: .no_toc }

Renders one of the active theme's button styles. Pick a `Variant`; the theme picks colors, padding, border, typography, and rounding.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Button struct {
    Variant   ButtonVariant
    Label     string
    Child     gutter.Widget
    OnPressed func()
}
```

| Field        | What it does                                                                                  |
| ------------ | --------------------------------------------------------------------------------------------- |
| `Variant`    | One of `ButtonPrimary`, `ButtonSecondary`, `ButtonGhost`, `ButtonAccent`, `ButtonOnDark`.     |
| `Label`      | Text label. Used if `Child` is nil.                                                           |
| `Child`      | Any widget as content (e.g. an icon + text Row). Overrides `Label` if set.                    |
| `OnPressed`  | Click handler. If nil, the button is visually a button but does nothing.                      |

---

## Variants

The same `Button{Variant: …}` produces very different visuals across themes:

| Variant            | Apple                                  | Meta                                                |
| ------------------ | -------------------------------------- | --------------------------------------------------- |
| `ButtonPrimary`    | Action Blue pill, 17px label           | Black pill, 17px / 700 label                        |
| `ButtonSecondary`  | Transparent + Action Blue outline pill | Transparent + black outline pill                    |
| `ButtonGhost`      | Pearl-fill 11px rounded, 14px label    | Very-soft-grey 100px pill, 14px / 700 label         |
| `ButtonAccent`     | Falls back to Action Blue              | Cobalt pill — the commerce buy CTA                  |
| `ButtonOnDark`     | Dark grey pill for use on light bg     | White pill with black label for dark surfaces       |

Use `ButtonAccent` for commerce-flow buys (`Add to cart`, `Pre-order`). Use `ButtonOnDark` whenever the button sits on a `SurfaceDark` or `CardPromo` background.

---

## Usage

```go
widgets.Button{
    Variant:   widgets.ButtonPrimary,
    Label:     "Get started",
    OnPressed: func() { startOnboarding() },
}

widgets.Button{
    Variant: widgets.ButtonGhost,
    Label:   "Cancel",
}
```

---

## A button with a custom child

When you need an icon-and-label, a custom-styled inner widget, or anything richer than plain text:

```go
widgets.Button{
    Variant: widgets.ButtonPrimary,
    Child: widgets.Row{
        Spacing:        8,
        CrossAxisAlign: widgets.CrossAxisCenter,
        Children: []gutter.Widget{
            widgets.Text{Data: "↓"},
            widgets.Text{Data: "Download"},
        },
    },
    OnPressed: func() { startDownload() },
}
```

When `Child` is set, `Label` is ignored.

---

## A row of CTAs

```go
widgets.Row{
    Spacing: 12,
    Children: []gutter.Widget{
        widgets.Button{Variant: widgets.ButtonAccent, Label: "Buy now", OnPressed: addToCart},
        widgets.Button{Variant: widgets.ButtonGhost, Label: "Learn more", OnPressed: showInfo},
    },
}
```

---

## A button on a dark surface

```go
widgets.Surface{
    Variant: widgets.SurfaceDark,
    Child: widgets.Button{
        Variant:   widgets.ButtonOnDark,
        Label:     "Get started",
        OnPressed: ctaPressed,
    },
}
```

---

## Notes

- Underlying tag is `<button>`. The `cursor: pointer` and `user-select: none` are applied by the widget.
- The button has a baked-in `transition: transform 0.15s, background-color 0.15s` — change it via [`Styled`](styled.html) if you need something else.
- `Button` doesn't disable itself when `OnPressed` is nil; the click just doesn't do anything. For a disabled state, wrap it in your own widget with `pointer-events: none`.

---

## See also

- [Themes](../themes.html) — `Components.ButtonPrimary` etc. and how to define a new theme's buttons.
- [`Link`](link.html) — for inline text actions, not buttons.
- [`Styled`](styled.html) — if you need a raw `<button>` with hand-rolled CSS.

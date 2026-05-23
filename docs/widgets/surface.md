---
title: Surface
parent: Widgets
nav_order: 9
---

# `Surface`
{: .no_toc }

A themed full-bleed region. Use it for hero bands, alternating sections, dark banners — anywhere you want a full-width tile with theme-driven background and padding.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Surface struct {
    Variant SurfaceVariant
    Padding string
    Child   gutter.Widget
}
```

| Field     | What it does                                                                                                                       |
| --------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `Variant` | One of `SurfaceCanvas`, `SurfaceAlt`, `SurfaceDark`.                                                                               |
| `Padding` | CSS shorthand override. Pass `"0"` for an edge-to-edge tile that controls its own child padding. Defaults to the theme's section padding. |
| `Child`   | Inner widget.                                                                                                                      |

---

## Variants

| Variant         | What it is                                                                  |
| --------------- | --------------------------------------------------------------------------- |
| `SurfaceCanvas` | The page-level canvas — white under both Apple and Meta.                    |
| `SurfaceAlt`    | The alternate light surface — parchment (`#f5f5f7`) on Apple, soft cloud on Meta. Used to break light tiles. |
| `SurfaceDark`   | The dark tile / banner — black on Apple, near-black on Meta.                |

---

## Sizing semantics

Surface has subtle but deliberate sizing:

- `width: 100%` — always full-width.
- `height: 100%` + `min-height: 100%` — fills its parent's available space; if content overflows, grows past the viewport.

For a Surface nested in an auto-height parent (like a `Column` of Surfaces), both percentages collapse and Surface sizes to its content — so vertically-stacked sections don't each force a full viewport.

For a Surface placed directly inside a `Scaffold.Body`, Body is `flex: 1` of a fixed-height flex column, so Surface fills the available viewport.

The net effect: drop a Surface into Body for a hero band, or stack several Surfaces inside a Column for a marketing page — both work the way you'd expect.

---

## Usage

```go
// Hero band
widgets.Surface{
    Variant: widgets.SurfaceDark,
    Child: widgets.Center{
        Child: widgets.Heading{
            Level: widgets.H1,
            Text:  "Hello",
            Color: ctx.Theme.Colors.OnDark,
        },
    },
}

// Alternating light section
widgets.Surface{
    Variant: widgets.SurfaceAlt,
    Child:   /* … */,
}
```

---

## A stacked marketing page

```go
widgets.Surface{
    Variant: widgets.SurfaceAlt,
    Padding: "0",
    Child: widgets.Column{
        Spacing: 0,
        Children: []gutter.Widget{
            widgets.Surface{Variant: widgets.SurfaceDark, Child: /* hero */},
            widgets.Surface{Variant: widgets.SurfaceCanvas, Child: /* feature row */},
            widgets.Surface{Variant: widgets.SurfaceAlt, Child: /* promo */},
            widgets.Surface{Variant: widgets.SurfaceCanvas, Child: /* badges */},
        },
    },
}
```

`Padding: "0"` on the outer Surface is the trick — it lets the inner Surfaces own their padding and edge-to-edge backgrounds.

---

## See also

- [`Card`](card.html) — bordered tile, not a region.
- [`Scaffold`](scaffold.html) — Body usually contains a Surface.
- [Themes](../themes.html) — `Components.SurfaceCanvas` / `SurfaceAlt` / `SurfaceDark`.

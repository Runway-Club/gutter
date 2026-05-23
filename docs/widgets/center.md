---
title: Center
parent: Widgets
nav_order: 13
---

# `Center`
{: .no_toc }

Centers a single child both horizontally and vertically inside a full-size box (`width: 100%`, `height: 100%`).
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Center struct {
    Child gutter.Widget
}
```

---

## Usage

```go
widgets.Center{
    Child: widgets.Heading{Level: widgets.H1, Text: "Hello"},
}
```

The `Center` itself is `display: flex; justify-content: center; align-items: center; width: 100%; height: 100%`. Drop it inside any sized parent and the child sits in the middle.

---

## Center inside a Scaffold body

The canonical landing-page pattern:

```go
widgets.Scaffold{
    AppBar: widgets.AppBar{Title: "Hello"},
    Body: widgets.Surface{
        Variant: widgets.SurfaceAlt,
        Child: widgets.Center{
            Child: widgets.Card{
                Variant: widgets.CardFeature,
                Child: widgets.Column{
                    CrossAxisAlign: widgets.CrossAxisCenter,
                    Spacing:        16,
                    Children: []gutter.Widget{
                        widgets.Heading{Level: widgets.H2, Text: "Hello"},
                        widgets.Button{Variant: widgets.ButtonPrimary, Label: "Start"},
                    },
                },
            },
        },
    },
}
```

`Scaffold` gives Body `flex: 1`, Surface fills it, Center centers inside Surface, Card shrink-wraps in the middle.

---

## Notes

- `Center` requires its parent to have a definite size. Inside an auto-sized parent, the `height: 100%` resolves against 0 and the center isn't visible. That's almost never an issue inside a `Scaffold`/`Surface` chain.
- For centering only horizontally (in a normal-flow column), use `Column` with `CrossAxisAlign: CrossAxisCenter` instead.

---

## See also

- [`Column`](column-row.html), [`Row`](column-row.html) — multi-child centering and alignment.
- [`Surface`](surface.html) — Center's typical parent.

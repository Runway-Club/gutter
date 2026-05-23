---
title: Card
parent: Widgets
nav_order: 8
---

# `Card`
{: .no_toc }

A themed bordered/filled box with one child. Use it for product tiles, feature cards, dark promo panels.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Card struct {
    Variant CardVariant
    Padding string
    Child   gutter.Widget
}
```

| Field     | What it does                                                                                |
| --------- | ------------------------------------------------------------------------------------------- |
| `Variant` | One of `CardFeature`, `CardPromo`, `CardPlain`.                                             |
| `Padding` | CSS shorthand override (e.g. `"32px"`, `"24px 32px"`). Defaults to the theme's card padding. |
| `Child`   | Inner widget (typically a `Column` or `Row`).                                               |

---

## Variants

| Variant       | What it is                                                                                  |
| ------------- | ------------------------------------------------------------------------------------------- |
| `CardFeature` | The bordered light card used for product/feature tiles. Default rounding, theme `Hairline` border. |
| `CardPromo`   | The dark promo surface — Meta's promo strip, Apple's full-bleed dark tile.                   |
| `CardPlain`   | A minimally-decorated rounded surface — no border, default padding.                          |

---

## Usage

```go
widgets.Card{
    Variant: widgets.CardFeature,
    Child: widgets.Column{
        Spacing: 8,
        Children: []gutter.Widget{
            widgets.Heading{Level: widgets.H5, Text: "Type"},
            widgets.Body{Text: "A complete typographic ladder.", Small: true},
        },
    },
}
```

---

## A row of feature cards

```go
widgets.Row{
    Spacing: 24,
    Children: []gutter.Widget{
        featureCard("Type", "From hero display to fine print."),
        featureCard("Color", "Every role mapped, no hex literals."),
        featureCard("Shape", "Pick a variant, the theme picks the values."),
    },
}

func featureCard(title, body string) gutter.Widget {
    return widgets.Card{
        Variant: widgets.CardFeature,
        Child: widgets.Column{
            Spacing: 8,
            Children: []gutter.Widget{
                widgets.Heading{Level: widgets.H5, Text: title},
                widgets.Body{Text: body, Small: true},
            },
        },
    }
}
```

---

## A dark promo card

```go
widgets.Card{
    Variant: widgets.CardPromo,
    Child: widgets.Column{
        CrossAxisAlign: widgets.CrossAxisCenter,
        Spacing:        16,
        Children: []gutter.Widget{
            widgets.Heading{
                Level: widgets.H3,
                Text:  "Ready to ship?",
                Color: ctx.Theme.Colors.OnDark,
            },
            widgets.Body{
                Text:  "Pre-order now and pay later.",
                Color: ctx.Theme.Colors.OnDark,
            },
            widgets.Button{Variant: widgets.ButtonAccent, Label: "Add to cart"},
        },
    },
}
```

On `CardPromo` you almost always want `Color: ctx.Theme.Colors.OnDark` on the typography widgets.

---

## Custom padding

```go
widgets.Card{
    Variant: widgets.CardFeature,
    Padding: "48px 32px",
    Child:   /* … */,
}
```

Pass any CSS shorthand. Omit to use the theme's default card padding.

---

## See also

- [`Surface`](surface.html) — when you want a full-bleed region, not a tile.
- [Themes](../themes.html) — `Components.CardFeature` / `CardPromo` / `CardPlain`.
- [`Column`](column-row.html), [`Row`](column-row.html) — the natural inner widgets.

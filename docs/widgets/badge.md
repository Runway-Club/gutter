---
title: Badge
parent: Widgets
nav_order: 11
---

# `Badge`
{: .no_toc }

A small status pill: `In stock`, `Limited time`, `Out of stock`, `New`.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Badge struct {
    Variant BadgeVariant
    Text    string
}
```

| Field     | What it does                                                          |
| --------- | --------------------------------------------------------------------- |
| `Variant` | One of `BadgeNeutral`, `BadgeSuccess`, `BadgeWarning`, `BadgeCritical`. |
| `Text`    | The pill label.                                                       |

Styling — background, foreground, rounding, padding, typography — comes from `theme.Components.Badge*`.

---

## Variants

| Variant         | Use for                                |
| --------------- | -------------------------------------- |
| `BadgeNeutral`  | Generic status (`In review`, `Draft`). |
| `BadgeSuccess`  | Positive (`In stock`, `Verified`).     |
| `BadgeWarning`  | Caution (`Selling fast`, `Limited`).   |
| `BadgeCritical` | Destructive / out (`Out of stock`).    |

---

## Usage

```go
widgets.Badge{Variant: widgets.BadgeSuccess, Text: "In stock"}
widgets.Badge{Variant: widgets.BadgeCritical, Text: "Out of stock"}
```

---

## A status row

```go
widgets.Row{
    Spacing:       12,
    MainAxisAlign: widgets.MainAxisCenter,
    Children: []gutter.Widget{
        widgets.Badge{Variant: widgets.BadgeNeutral, Text: "In review"},
        widgets.Badge{Variant: widgets.BadgeSuccess, Text: "In stock"},
        widgets.Badge{Variant: widgets.BadgeWarning, Text: "Selling fast"},
        widgets.Badge{Variant: widgets.BadgeCritical, Text: "Out of stock"},
    },
}
```

---

## Notes

- `Badge` is non-interactive. For a clickable chip, wrap it in [`Styled`](styled.html) with an event handler, or use [`Button{Variant: ButtonGhost}`](button.html) with smaller typography.
- The pill is `inline-flex` and shrink-wraps to its content.

---

## See also

- [Themes](../themes.html) — `Components.BadgeNeutral` etc.

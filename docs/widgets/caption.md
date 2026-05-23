---
title: Caption
parent: Widgets
nav_order: 5
---

# `Caption`
{: .no_toc }

Shorthand for [`Body{Small: true}`](body.html) — caption-sized text in the theme's `Caption` typography role.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Caption struct {
    Text  string
    Bold  bool
    Color string
}
```

| Field   | What it does                                                          |
| ------- | --------------------------------------------------------------------- |
| `Text`  | The string to render.                                                 |
| `Bold`  | If true, use `Typography.CaptionStrong` instead of `Typography.Caption`. |
| `Color` | Optional override. Defaults to `ctx.Theme.Colors.Ink`.                |

`Caption` is literally implemented as `Body{Small: true}` plus the fields above.

---

## Usage

```go
widgets.Caption{Text: "Tap to learn more"}
widgets.Caption{Text: "Required", Bold: true}
widgets.Caption{Text: "Disabled", Color: ctx.Theme.Colors.InkSubtle}
```

Use Caption for fine-print, table-cell secondary lines, footer legal text, form field hints.

---

## See also

- [`Body`](body.html) — the underlying widget.
- [`Heading`](heading.html) — for larger sizes.

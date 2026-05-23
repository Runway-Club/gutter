---
title: Padding
parent: Widgets
nav_order: 14
---

# `Padding`
{: .no_toc }

Wraps a child in a `<div>` with the given padding. Use [`EdgeInsets`](#edgeinsets) to describe symmetric or per-side padding.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Padding struct {
    Padding EdgeInsets
    Child   gutter.Widget
}
```

---

## EdgeInsets

```go
type EdgeInsets struct {
    Top, Right, Bottom, Left float64 // pixels
}

func EdgeInsetsAll(v float64) EdgeInsets
func EdgeInsetsSymmetric(vertical, horizontal float64) EdgeInsets
```

Both helpers return an `EdgeInsets` you can pass directly:

```go
widgets.EdgeInsetsAll(16)               // 16 on every side
widgets.EdgeInsetsSymmetric(24, 32)     // 24 top/bottom, 32 left/right
widgets.EdgeInsets{Top: 8, Left: 12}    // per-side, others default to 0
```

---

## Usage

```go
widgets.Padding{
    Padding: widgets.EdgeInsetsAll(16),
    Child:   widgets.Heading{Level: widgets.H4, Text: "Hello"},
}
```

```go
widgets.Padding{
    Padding: widgets.EdgeInsetsSymmetric(24, 32),
    Child: widgets.Column{
        Spacing: 12,
        Children: []gutter.Widget{
            widgets.Heading{Level: widgets.H4, Text: "Settings"},
            widgets.Body{Text: "Adjust your preferences below."},
        },
    },
}
```

---

## Notes

- `Padding` is purely structural — it doesn't apply background color, border, or anything else. For padding **plus** a colored box, use [`Container`](container.html) or [`Card`](card.html) instead.
- Internally, `EdgeInsets.CSS()` produces the `top right bottom left` shorthand string assigned to the wrapper's `padding`.

---

## See also

- [`Container`](container.html) — when you also want background / border / radius.
- [`SizedBox`](sizedbox.html) — when you want a fixed-size spacer instead of padding.

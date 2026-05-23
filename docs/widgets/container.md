---
title: Container
parent: Widgets
nav_order: 16
---

# `Container`
{: .no_toc }

A styled `<div>` with a single optional child. Use it as a general-purpose layout box for background color, padding, sizing, border, and radius — when the themed widgets don't cover what you need.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Container struct {
    Child        gutter.Widget
    Padding      EdgeInsets
    Margin       EdgeInsets
    Color        string  // background-color
    Width        string  // any CSS length
    Height       string
    BorderRadius string
    Border       string  // full CSS border shorthand, e.g. "1px solid #ccc"
}
```

All fields are optional. Empty values are not written to the style — a `Container{}` produces a stripped-down `<div>`.

---

## Usage

```go
widgets.Container{
    Color:        "#0066cc",
    Padding:      widgets.EdgeInsetsAll(24),
    BorderRadius: "12px",
    Child: widgets.Text{
        Data:  "Hello on blue",
        Style: &widgets.TextStyle{Color: "#ffffff"},
    },
}
```

```go
widgets.Container{
    Width:  "320px",
    Height: "180px",
    Color:  "#f5f5f7",
    Border: "1px solid #e0e0e0",
}
```

---

## When to use Container vs Card vs Surface

| If you need…                                             | Use…                              |
| -------------------------------------------------------- | --------------------------------- |
| A themed product/feature tile with theme colors          | [`Card`](card.html)               |
| A themed full-bleed region (hero, banner, alt section)   | [`Surface`](surface.html)         |
| A raw-CSS box with arbitrary colors / borders / sizing   | `Container`                       |
| Just padding, no background                              | [`Padding`](padding.html)         |
| Just sizing, no color                                    | [`SizedBox`](sizedbox.html)       |

Container is the right tool when you're building a custom widget that doesn't fit any theme variant — for example, a colored swatch, a divider, a placeholder rectangle for an image you haven't loaded yet.

---

## Notes

- `Container` is a HostWidget (it implements `Host()` directly), not a StatelessWidget. There's no `Build` call — the runtime mounts the `<div>` immediately.
- `Color` is **background-color**, not text color. For text color, set it on the child widget (e.g. `TextStyle.Color`).

---

## See also

- [`EdgeInsets`](padding.html#edgeinsets) — same type used by `Padding` / `Margin` here.
- [`Styled`](styled.html) — if you need a tag other than `<div>` or arbitrary CSS properties.

---
title: SizedBox
parent: Widgets
nav_order: 15
---

# `SizedBox`
{: .no_toc }

Forces a fixed width and/or height. Empty values let CSS pick. Use it as a spacer, a measured container, or a fixed slot.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type SizedBox struct {
    Width  string  // any CSS length: "120px", "50%", "10rem"
    Height string
    Child  gutter.Widget
}
```

Both `Width` and `Height` are CSS strings — pass `"120px"`, `"50%"`, `"min(100%, 400px)"`, or whatever you need.

---

## Usage

### A vertical spacer

```go
widgets.Column{
    Children: []gutter.Widget{
        widgets.Heading{Level: widgets.H2, Text: "Hello"},
        widgets.SizedBox{Height: "32px"},
        widgets.Body{Text: "World"},
    },
}
```

Equivalent to a `Column{Spacing: 32}` if every gap is the same; reach for `SizedBox` when you need one specific gap different from the rest.

### A horizontal spacer in a Row

```go
widgets.Row{
    Children: []gutter.Widget{
        widgets.Text{Data: "Label"},
        widgets.SizedBox{Width: "16px"},
        widgets.Text{Data: "Value"},
    },
}
```

### A fixed slot

```go
widgets.SizedBox{
    Width:  "240px",
    Height: "180px",
    Child:  widgets.Container{Color: "#0066cc"},
}
```

---

## Notes

- Both `Width` and `Height` are optional; an empty value just isn't written to the style.
- Without a `Child`, `SizedBox` is a useful spacer or placeholder.
- For padding around an existing widget instead of a fixed-size box, see [`Padding`](padding.html).

---

## See also

- [`Padding`](padding.html), [`Container`](container.html) — related layout primitives.

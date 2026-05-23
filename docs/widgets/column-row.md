---
title: Column & Row
parent: Widgets
nav_order: 12
---

# `Column` and `Row`
{: .no_toc }

Flex layouts. `Column` lays children out vertically; `Row` lays them out horizontally. Both expose the same axis-alignment knobs and a gap-style `Spacing`.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Column struct {
    Children       []gutter.Widget
    MainAxisAlign  string  // CSS justify-content
    CrossAxisAlign string  // CSS align-items
    Spacing        float64 // CSS gap in pixels
}

type Row struct {
    Children       []gutter.Widget
    MainAxisAlign  string
    CrossAxisAlign string
    Spacing        float64
}
```

Both are HostWidgets — they map directly to a flex `<div>`.

| Field            | What it does                                                                  |
| ---------------- | ----------------------------------------------------------------------------- |
| `Children`       | Children laid out along the main axis.                                        |
| `MainAxisAlign`  | Alignment along the main axis (vertical for `Column`, horizontal for `Row`).  |
| `CrossAxisAlign` | Alignment perpendicular to the main axis.                                     |
| `Spacing`        | Gap between children in pixels. Maps to CSS `gap`.                            |

---

## Alignment constants

Use the constants (they match the underlying CSS keywords):

```go
const (
    MainAxisStart        = "flex-start"
    MainAxisCenter       = "center"
    MainAxisEnd          = "flex-end"
    MainAxisSpaceBetween = "space-between"
    MainAxisSpaceAround  = "space-around"
    MainAxisSpaceEvenly  = "space-evenly"
)

const (
    CrossAxisStart    = "flex-start"
    CrossAxisCenter   = "center"
    CrossAxisEnd      = "flex-end"
    CrossAxisStretch  = "stretch"
    CrossAxisBaseline = "baseline"
)
```

Raw CSS keywords also work — `"center"`, `"flex-end"` — but prefer the constants for clarity.

---

## Usage

### A vertical stack

```go
widgets.Column{
    Spacing: 16,
    Children: []gutter.Widget{
        widgets.Heading{Level: widgets.H2, Text: "Hello"},
        widgets.Body{Text: "Welcome to Gutter."},
        widgets.Button{Variant: widgets.ButtonPrimary, Label: "Get started"},
    },
}
```

### A horizontal row of buttons

```go
widgets.Row{
    Spacing: 12,
    Children: []gutter.Widget{
        widgets.Button{Variant: widgets.ButtonGhost, Label: "Cancel"},
        widgets.Button{Variant: widgets.ButtonPrimary, Label: "Save"},
    },
}
```

### Centered vertically + horizontally

```go
widgets.Column{
    MainAxisAlign:  widgets.MainAxisCenter,
    CrossAxisAlign: widgets.CrossAxisCenter,
    Spacing:        16,
    Children:       []gutter.Widget{ /* … */ },
}
```

This places children in the middle vertically (main axis) and centers each one horizontally (cross axis).

### Space-between layout (e.g. header row)

```go
widgets.Row{
    MainAxisAlign:  widgets.MainAxisSpaceBetween,
    CrossAxisAlign: widgets.CrossAxisCenter,
    Children: []gutter.Widget{
        widgets.Heading{Level: widgets.H4, Text: "Inbox"},
        widgets.Button{Variant: widgets.ButtonPrimary, Label: "Compose"},
    },
}
```

---

## Stretching children to fill

`CrossAxisStretch` makes children fill the cross axis. Useful for tabular layouts:

```go
widgets.Row{
    CrossAxisAlign: widgets.CrossAxisStretch,
    Children: []gutter.Widget{
        sidebar,
        mainContent,
    },
}
```

To make a specific child grow to fill the main axis, wrap it in a `Styled` with `flex: 1`:

```go
widgets.Row{
    Children: []gutter.Widget{
        widgets.Text{Data: "Label"},
        widgets.Styled{Style: map[string]string{"flex": "1"}}, // spacer
        widgets.Button{Variant: widgets.ButtonGhost, Label: "Action"},
    },
}
```

---

## Notes

- The `<div>` Column/Row produce always has `display: flex` plus a `flex-direction` of `column` or `row`.
- `Spacing` translates to CSS `gap`, which is supported in every modern browser.
- For a centered single-child layout, prefer [`Center`](center.html) over `Column` with axis alignment — Center is more specific and clearer at the call site.

---

## See also

- [`Center`](center.html) — center a single child in a full-size box.
- [`Padding`](padding.html), [`SizedBox`](sizedbox.html) — composition helpers.

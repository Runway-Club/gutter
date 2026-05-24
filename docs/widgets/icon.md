---
title: Icon
parent: Widgets
nav_order: 29
---

# `Icon`
{: .no_toc }

A Google Material Symbols glyph, rendered as a `<span>` element with the right variable-font axes set. Pair with [`IconButton`](iconbutton.html) for tappable icons.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Icon struct {
    Name   string
    Size   string
    Color  string
    Style  IconStyle // IconOutlined (default) | IconRounded | IconSharp
    Filled bool
    Weight int // 100..700 (default 400)
    Grade  int // -25..200
}
```

| Field    | What it does                                                                      |
| -------- | --------------------------------------------------------------------------------- |
| `Name`   | Symbol identifier (lowercase snake_case): `"home"`, `"favorite"`, `"arrow_back"`. |
| `Size`   | CSS size. Defaults to `"24px"`.                                                   |
| `Color`  | Optional explicit color. Defaults to `currentColor` (inherits text color).        |
| `Style`  | Pick the family — outlined, rounded, sharp.                                       |
| `Filled` | Toggle the FILL axis on the variable font.                                        |
| `Weight` | Stroke weight axis (100..700). Defaults to 400.                                   |
| `Grade`  | Optical-grade bias axis (-25..200). Defaults to 0.                                |

Browse the full glyph catalog at [fonts.google.com/icons](https://fonts.google.com/icons).

---

## Page setup

The hosting page must include the Material Symbols stylesheet(s). The scaffolded `index.html` from `gutter new` already preloads all three families with the full axis range:

```html
<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200" rel="stylesheet">
<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Rounded:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200" rel="stylesheet">
<link href="https://fonts.googleapis.com/css2?family=Material+Symbols+Sharp:opsz,wght,FILL,GRAD@20..48,100..700,0..1,-50..200" rel="stylesheet">
```

Drop any family you don't use.

---

## Usage

```go
widgets.Icon{Name: "home"}                            // default 24px outlined
widgets.Icon{Name: "favorite", Filled: true, Color: "#e91e63"}
widgets.Icon{Name: "settings", Style: widgets.IconRounded, Weight: 600}
widgets.Icon{Name: "star", Style: widgets.IconSharp, Size: "32px"}
```

Inside a row of icons + labels:

```go
widgets.Row{
    CrossAxisAlign: widgets.CrossAxisCenter,
    Spacing:        8,
    Children: []gutter.Widget{
        widgets.Icon{Name: "check_circle", Color: "#34c759"},
        widgets.Body{Text: "Saved"},
    },
}
```

---

## Notes

- The `opsz` axis is derived from `Size` so the optical-size automatically tracks the rendered glyph size.
- For interactive icons use [`IconButton`](iconbutton.html), which handles padding, hover, and aria-label.

---

## See also

- [IconButton](iconbutton.html) — tappable icon with a Button-style background.
- [Image](image.html) — for raster or arbitrary SVG content.

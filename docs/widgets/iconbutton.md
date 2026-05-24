---
title: IconButton
parent: Widgets
nav_order: 30
---

# `IconButton`
{: .no_toc }

A square Button whose only content is an [`Icon`](icon.html). Pulls the same theme palette as `Button` — `Variant` picks `ButtonPrimary/Secondary/Ghost/Accent/OnDark` — but switches to pill-shaped square padding.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type IconButton struct {
    Icon      string
    IconStyle IconStyle
    Filled    bool
    Size      string
    Tooltip   string
    Variant   ButtonVariant
    OnPressed func()
}
```

| Field       | What it does                                                          |
| ----------- | --------------------------------------------------------------------- |
| `Icon`      | Material Symbols glyph name.                                          |
| `IconStyle` | Pick the symbol family. Defaults to outlined.                         |
| `Filled`    | Toggle the FILL axis on the icon.                                     |
| `Size`      | Icon size CSS value. Defaults to `"24px"`.                            |
| `Tooltip`   | Exposed as both `title` and `aria-label` for accessibility.           |
| `Variant`   | Button palette: `ButtonPrimary`, `ButtonGhost`, etc.                  |
| `OnPressed` | Click callback.                                                       |

---

## Usage

```go
widgets.IconButton{Icon: "favorite", Variant: widgets.ButtonGhost, Filled: true, Tooltip: "Like"}
widgets.IconButton{Icon: "delete", Variant: widgets.ButtonGhost, IconStyle: widgets.IconRounded}
widgets.IconButton{Icon: "bookmark", Variant: widgets.ButtonPrimary, Tooltip: "Save"}
```

In an AppBar:

```go
widgets.AppBar{
    Title: "Gutter",
    Leading: widgets.IconButton{Icon: "menu", Variant: widgets.ButtonGhost, OnPressed: openDrawer},
    Actions: []gutter.Widget{
        widgets.IconButton{Icon: "search", Variant: widgets.ButtonGhost, Tooltip: "Search"},
        widgets.IconButton{Icon: "settings", Variant: widgets.ButtonGhost, Tooltip: "Settings"},
    },
}
```

---

## See also

- [Icon](icon.html) — the underlying glyph widget.
- [Button](button.html) — the themed button this widget composes on top of.

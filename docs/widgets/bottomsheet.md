---
title: BottomSheet
parent: Widgets
nav_order: 42
---

# `BottomSheet`
{: .no_toc }

A panel that slides up from the bottom edge of the viewport. Same `Listenable[bool]` visibility pattern as [`Popup`](popup.html) and [`Drawer`](drawer.html).
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type BottomSheet struct {
    Open      gutter.Listenable[bool]
    Child     gutter.Widget
    Height    string
    OnDismiss func()
    ZIndex    string
}
```

| Field       | What it does                                                              |
| ----------- | ------------------------------------------------------------------------- |
| `Open`      | Visibility source.                                                        |
| `Child`     | Content inside the sheet.                                                 |
| `Height`    | Maximum sheet height. Defaults to `"min(60vh, 480px)"`.                   |
| `OnDismiss` | Called on backdrop click.                                                 |
| `ZIndex`    | Defaults to `"1000"`.                                                     |

---

## Basic usage

```go
sheetOpen := gutter.NewNotifier(false)

widgets.IconButton{Icon: "more_vert", OnPressed: func() { sheetOpen.Set(true) }}

widgets.BottomSheet{
    Open:      sheetOpen,
    OnDismiss: func() { sheetOpen.Set(false) },
    Child: widgets.Column{
        Spacing: 8,
        Children: []gutter.Widget{
            widgets.Heading{Level: widgets.H4, Text: "Quick actions"},
            widgets.Button{Variant: widgets.ButtonGhost, Label: "Share"},
            widgets.Button{Variant: widgets.ButtonGhost, Label: "Duplicate"},
            widgets.Button{Variant: widgets.ButtonGhost, Label: "Archive"},
        },
    },
}
```

---

## See also

- [Popup](popup.html), [Drawer](drawer.html) — the other overlay variants.

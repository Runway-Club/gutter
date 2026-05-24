---
title: Drawer
parent: Widgets
nav_order: 41
---

# `Drawer`
{: .no_toc }

A side panel that slides in from the left or right edge of the viewport. Visibility is driven by a `Listenable[bool]`, mirroring [`Popup`](popup.html).
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type DrawerSide int
const ( DrawerLeft DrawerSide = iota; DrawerRight )

type Drawer struct {
    Open      gutter.Listenable[bool]
    Child     gutter.Widget
    Side      DrawerSide
    Width     string
    OnDismiss func()
    ZIndex    string
}
```

| Field       | What it does                                                                  |
| ----------- | ----------------------------------------------------------------------------- |
| `Open`      | Visibility source.                                                            |
| `Child`     | Content inside the drawer panel.                                              |
| `Side`      | `DrawerLeft` (default) or `DrawerRight`.                                      |
| `Width`     | Panel width. Defaults to `"min(80vw, 320px)"`.                                |
| `OnDismiss` | Called on backdrop click.                                                     |
| `ZIndex`    | Defaults to `"1000"`.                                                         |

---

## Basic usage

```go
drawerOpen := gutter.NewNotifier(false)

widgets.Scaffold{
    AppBar: widgets.AppBar{
        Leading: widgets.IconButton{Icon: "menu", OnPressed: func() { drawerOpen.Set(true) }},
        Title:   "My App",
    },
    Body: /* page content */,
}

// Drawer can sit anywhere — its position:fixed means it doesn't occupy
// layout space. Put it inside Scaffold.Body's column for tidiness.
widgets.Drawer{
    Open:      drawerOpen,
    Side:      widgets.DrawerLeft,
    OnDismiss: func() { drawerOpen.Set(false) },
    Child: widgets.Column{
        Spacing: 16,
        Children: []gutter.Widget{
            widgets.Heading{Level: widgets.H4, Text: "Menu"},
            widgets.Link{Text: "Home"},
            widgets.Link{Text: "Settings"},
        },
    },
}
```

---

## See also

- [Popup](popup.html), [BottomSheet](bottomsheet.html) — the other overlay variants.

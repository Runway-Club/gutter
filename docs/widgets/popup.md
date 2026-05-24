---
title: Popup
parent: Widgets
nav_order: 40
---

# `Popup`
{: .no_toc }

A centered modal dialog with a dim backdrop. Visibility is driven by a `Listenable[bool]` — typically a `Notifier` the app holds and flips from a button — so the open/closed state lives in app code and the popup is just a rendering of it.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Popup struct {
    Open      gutter.Listenable[bool]
    Child     gutter.Widget
    OnDismiss func()
    MaxWidth  string
    ZIndex    string
}
```

| Field       | What it does                                                                                |
| ----------- | ------------------------------------------------------------------------------------------- |
| `Open`      | Visibility source. The widget observes it and rebuilds when it fires.                       |
| `Child`     | Content inside the popup card.                                                              |
| `OnDismiss` | Called when the user clicks the backdrop. Leave nil for a non-dismissible popup.            |
| `MaxWidth`  | Caps the sheet's width. Defaults to `"min(90vw, 480px)"`.                                   |
| `ZIndex`    | Base z-index for the overlay layer. Defaults to `"1000"`.                                   |

---

## Basic usage

```go
open := gutter.NewNotifier(false)

widgets.Column{
    Children: []gutter.Widget{
        widgets.Button{
            Label:     "Open dialog",
            OnPressed: func() { open.Set(true) },
        },
        widgets.Popup{
            Open:      open,
            OnDismiss: func() { open.Set(false) },
            Child: widgets.Column{
                Spacing: 12,
                Children: []gutter.Widget{
                    widgets.Heading{Level: widgets.H4, Text: "Are you sure?"},
                    widgets.Body{Text: "This will delete your account."},
                    widgets.Row{
                        Spacing: 8,
                        Children: []gutter.Widget{
                            widgets.Button{Variant: widgets.ButtonGhost, Label: "Cancel",
                                OnPressed: func() { open.Set(false) }},
                            widgets.Button{Variant: widgets.ButtonPrimary, Label: "Delete",
                                OnPressed: func() { /* delete + close */ open.Set(false) }},
                        },
                    },
                },
            },
        },
    },
}
```

---

## Implementation notes

- The popup is always mounted; the open/closed states differ only in CSS so the fade+scale transition runs in both directions.
- Backdrop and content are siblings under a `display: contents` wrapper, not parent/child — so clicks on the content don't bubble into the backdrop dismiss handler.
- Z-index defaults to 1000. If you stack multiple overlays, bump explicitly.

---

## See also

- [Drawer](drawer.html), [BottomSheet](bottomsheet.html) — the other overlay variants.
- [Notifier + ObserverBuilder](observerbuilder.html) — the reactive plumbing the overlay subscribes to.

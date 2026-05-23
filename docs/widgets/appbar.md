---
title: AppBar
parent: Widgets
nav_order: 2
---

# `AppBar`
{: .no_toc }

The top navigation strip used inside a [`Scaffold`](scaffold.html). Layout: `[Leading] [Title] … spacer … [Actions]`.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type AppBar struct {
    Title       string
    TitleWidget gutter.Widget
    Leading     gutter.Widget
    Actions     []gutter.Widget
}
```

| Field         | What it does                                                                   |
| ------------- | ------------------------------------------------------------------------------ |
| `Title`       | Plain-text title. Styled with the theme's NavBar typography.                   |
| `TitleWidget` | Any widget as title — overrides `Title` if both are set.                       |
| `Leading`     | Leading-edge widget (icon, back arrow, logo).                                  |
| `Actions`     | Trailing widgets pushed to the right edge by a flex spacer.                    |

Background, height, padding, typography, and bottom border come from `theme.Components.NavBar`. App code never sets CSS on AppBar.

---

## Basic usage

```go
widgets.AppBar{Title: "Hello"}
```

Renders the title left-aligned, in the theme's NavBar foreground color and typography.

---

## With actions

Actions sit on the right edge, separated from the title by a flex spacer:

```go
widgets.AppBar{
    Title: "Inbox",
    Actions: []gutter.Widget{
        widgets.Button{Variant: widgets.ButtonGhost, Label: "Settings"},
        widgets.Button{Variant: widgets.ButtonPrimary, Label: "Compose"},
    },
}
```

The actions are rendered in order; the rightmost button visually sits at the right edge.

---

## With a custom title widget

Use `TitleWidget` when you need a logo, a brand wordmark, or any layout more elaborate than a single string:

```go
widgets.AppBar{
    TitleWidget: widgets.Row{
        Spacing:        8,
        CrossAxisAlign: widgets.CrossAxisCenter,
        Children: []gutter.Widget{
            widgets.Container{Width: "24px", Height: "24px", Color: "#0066cc"},
            widgets.Text{Data: "Acme", Style: &widgets.TextStyle{FontWeight: "600"}},
        },
    },
    Actions: []gutter.Widget{
        widgets.Button{Variant: widgets.ButtonGhost, Label: "Docs"},
    },
}
```

If both `TitleWidget` and `Title` are set, `TitleWidget` wins.

---

## With a leading widget

Leading widgets sit on the left edge, before the title:

```go
widgets.AppBar{
    Leading: widgets.Button{
        Variant:   widgets.ButtonGhost,
        Label:     "←",
        OnPressed: navigateBack,
    },
    Title: "Settings",
}
```

---

## Theme-specific looks

| Theme           | NavBar look                                                                 |
| --------------- | --------------------------------------------------------------------------- |
| `themes.Apple`  | 44px pure-black bar, 12px white labels with `-0.12px` tracking.             |
| `themes.Meta`   | 64px white bar with hairline-soft bottom border, 14px / 700 button-style label. |
| `themes.Neutral`| Light grey bar with system-font labels.                                     |

If you need to customize the bar globally, edit your theme's `Components.NavBar` (a `themes.NavBarStyle`) rather than the AppBar widget itself.

---

## See also

- [`Scaffold`](scaffold.html) — pair `AppBar` with `Scaffold.AppBar`.
- [`Button`](button.html) — common content for `Actions`.
- [Themes](../themes.html) — the `NavBarStyle` AppBar reads from.

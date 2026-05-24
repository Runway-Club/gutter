---
title: TextArea
parent: Widgets
nav_order: 21
---

# `TextArea`
{: .no_toc }

Multi-line themed text field. Reuses the theme's input styling for background/border/typography, with a softer corner radius for the larger surface.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type TextArea struct {
    Value       string
    Placeholder string
    Rows        int
    Error       bool
    Disabled    bool
    ReadOnly    bool
    OnChanged   func(string)

    Resize    string // "none" | "both" | "horizontal" | "vertical" (default)
    MaxLength int
    Name      string
}
```

| Field         | What it does                                                                              |
| ------------- | ----------------------------------------------------------------------------------------- |
| `Value`       | Controlled value. Set via the DOM `value` property (caret-preserving).                    |
| `Placeholder` | Placeholder shown when empty.                                                             |
| `Rows`        | Initial row count. Defaults to 4.                                                         |
| `Error`       | Use the theme's error border color.                                                       |
| `Disabled` / `ReadOnly` | Toggle interaction.                                                            |
| `OnChanged`   | Fires on the DOM `input` event with the new full value.                                   |
| `Resize`      | CSS `resize` value. Defaults to `"vertical"`.                                             |
| `MaxLength`   | Caps the number of characters the user can type.                                          |
| `Name`        | Form field name for native form submission.                                               |

The border radius uses the theme's `Rounded.Large` (e.g. 18px on Apple) rather than the pill-shaped `Components.Input.Rounded` so multi-line surfaces don't look strange.

---

## Basic usage

```go
widgets.TextArea{
    Value:       s.bio,
    Placeholder: "Tell us about yourself…",
    Rows:        5,
    MaxLength:   500,
    OnChanged:   func(v string) { s.SetState(func() { s.bio = v }) },
}
```

For a character counter:

```go
widgets.Column{
    Spacing: 4,
    Children: []gutter.Widget{
        widgets.TextArea{ /* … */ },
        widgets.Caption{Text: fmt.Sprintf("%d / 500 characters", len(s.bio))},
    },
}
```

---

## Notes

- The value sync uses the DOM `value` property only when the new value differs from what's already in the DOM — typing keeps the caret position even though the parent rebuilds on every keystroke.
- For single-line input use [`Input`](input.html).

---

## See also

- [Input](input.html) — single-line counterpart.
- [State Management](../state-management.html).

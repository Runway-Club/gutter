---
title: Select
parent: Widgets
nav_order: 25
---

# `Select[T comparable]`
{: .no_toc }

A themed `<select>` dropdown over a strongly-typed option set. The HTML option value is the slice index, so any comparable `T` works — strings, ints, custom structs — without needing a Stringer.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type SelectOption[T any] struct {
    Value T
    Label string
}

type Select[T comparable] struct {
    Options     []SelectOption[T]
    Selected    T
    OnChanged   func(T)
    Disabled    bool
    Placeholder string
    Name        string
}
```

| Field         | What it does                                                                              |
| ------------- | ----------------------------------------------------------------------------------------- |
| `Options`     | The choices. `Label` is shown to the user; `Value` is returned via `OnChanged`.           |
| `Selected`    | Source of truth. Synced into the DOM `value` property on every rebuild.                   |
| `OnChanged`   | Fires with the picked option's typed value.                                               |
| `Disabled`    | Disables the dropdown.                                                                    |
| `Placeholder` | When non-empty, prepends a disabled option with this text shown if nothing is selected.   |
| `Name`        | Form field name.                                                                          |

The dropdown reuses `theme.Components.Input` styling for the box; the visible value is centered (`text-align-last: center`) and the line-height is normalized so the glyph sits centered next to the native chevron.

---

## Basic usage

```go
type Color int
const ( Red Color = iota; Green; Blue )

widgets.Select[Color]{
    Options: []widgets.SelectOption[Color]{
        {Value: Red,   Label: "Red"},
        {Value: Green, Label: "Green"},
        {Value: Blue,  Label: "Blue"},
    },
    Selected:  s.current,
    OnChanged: func(c Color) { s.SetState(func() { s.current = c }) },
}
```

---

## With a placeholder

```go
widgets.Select[string]{
    Placeholder: "Pick a country",
    Options: []widgets.SelectOption[string]{
        {Value: "vn", Label: "Vietnam"},
        {Value: "us", Label: "United States"},
        {Value: "jp", Label: "Japan"},
    },
    Selected:  s.country,
    OnChanged: func(c string) { s.SetState(func() { s.country = c }) },
}
```

The placeholder shows until the user picks anything, then disappears.

---

## Notes

- The `selected` attribute on `<option>` is only honored at parse time — once the user picks, setAttribute doesn't update the rendered selection. Select sets the DOM `value` property directly in OnMount to keep the displayed choice in sync with `Selected`.

---

## See also

- [RadioGroup](radiogroup.html) — the visually-expanded equivalent.

---
title: Slider
parent: Widgets
nav_order: 24
---

# `Slider`
{: .no_toc }

A horizontal range input rendered as a native `<input type="range">` — themed via the CSS `accent-color` property.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Slider struct {
    Value     float64
    Min       float64
    Max       float64
    Step      float64
    Disabled  bool
    OnChanged func(float64)
    Name      string
}
```

| Field       | What it does                                                                  |
| ----------- | ----------------------------------------------------------------------------- |
| `Value`     | Current value (also synced into the DOM `value` property on every rebuild).   |
| `Min/Max`   | Range bounds. Defaults to 0..100 when both are zero.                          |
| `Step`      | Increment. Defaults to 1.                                                     |
| `Disabled`  | Disables interaction.                                                         |
| `OnChanged` | Fires for every drag tick (the DOM `input` event) with the new float value.   |
| `Name`      | Form field name.                                                              |

---

## Basic usage

```go
widgets.Slider{
    Value:     s.volume,
    Min:       0,
    Max:       100,
    Step:      1,
    OnChanged: func(v float64) { s.SetState(func() { s.volume = v }) },
}
```

For a labeled slider with a live value readout:

```go
widgets.Column{
    Spacing: 4,
    Children: []gutter.Widget{
        widgets.Caption{Text: fmt.Sprintf("Volume: %.0f", s.volume)},
        widgets.Slider{ /* … */ },
    },
}
```

---

## Notes

- `Value` is parsed from the DOM event's string and dispatched as a `float64` — non-numeric values are silently dropped.
- The thumb and track colors follow `accent-color` (theme primary), so they re-skin with the theme.

---

## See also

- [Themes](../themes.html) — `Colors.Primary` drives the accent color.

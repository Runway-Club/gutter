---
title: Switch
parent: Widgets
nav_order: 23
---

# `Switch`
{: .no_toc }

A two-state toggle styled as a sliding pill. Visually distinct from `Checkbox` — better for settings panels where the state reads more naturally as "on/off" than "checked/unchecked".
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Switch struct {
    Checked   bool
    Label     string
    Disabled  bool
    OnChanged func(bool)
}
```

| Field       | What it does                                                          |
| ----------- | --------------------------------------------------------------------- |
| `Checked`   | Source of truth.                                                      |
| `Label`     | Optional text rendered next to the toggle.                            |
| `Disabled`  | Disables interaction.                                                 |
| `OnChanged` | Fires with the new boolean on click.                                  |

---

## Basic usage

```go
widgets.Switch{
    Checked:   s.notifyMe,
    Label:     "Send weekly digest",
    OnChanged: func(v bool) { s.SetState(func() { s.notifyMe = v }) },
}
```

---

## Implementation notes

The widget is a `<button role="switch">` with a sibling thumb absolutely positioned inside the track. Track color comes from `theme.Colors.Hairline` (off) and `theme.Colors.Primary` (on). The transitions are pure CSS — no animation framework needed for the slide.

A real native checkbox isn't used because CSS `:checked +` sibling selectors only work with stylesheets, and gutter doesn't inject one — every style is inline.

---

## See also

- [Checkbox](checkbox.html) — native checkbox-style toggle.

---
title: RadioGroup
parent: Widgets
nav_order: 26
---

# `RadioGroup[T comparable]`
{: .no_toc }

A set of native `<input type="radio">` buttons sharing a `name` so the browser enforces single-selection. Generic over the option value type — works with typed enums or domain values rather than raw strings.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type RadioOption[T any] struct {
    Value T
    Label string
}

type RadioGroup[T comparable] struct {
    Options   []RadioOption[T]
    Selected  T
    OnChanged func(T)
    Disabled  bool
    Direction string // "column" (default) | "row"
}
```

| Field       | What it does                                                                              |
| ----------- | ----------------------------------------------------------------------------------------- |
| `Options`   | Choices, each labeled and typed.                                                          |
| `Selected`  | Source of truth — which option is currently picked.                                       |
| `OnChanged` | Fires with the picked option's typed value.                                               |
| `Disabled`  | Disables the whole group.                                                                 |
| `Direction` | Layout axis. `"row"` for horizontal pills, `"column"` for stacked. Defaults to column.    |

Each radio is themed via `accent-color` and laid out next to its label inside a `<label>` so clicks toggle the input.

---

## Basic usage

```go
widgets.RadioGroup[string]{
    Direction: "row",
    Options: []widgets.RadioOption[string]{
        {Value: "s", Label: "Small"},
        {Value: "m", Label: "Medium"},
        {Value: "l", Label: "Large"},
    },
    Selected:  s.size,
    OnChanged: func(v string) { s.SetState(func() { s.size = v }) },
}
```

---

## Multiple groups on one page

The widget is StatefulWidget-backed: on mount it generates a unique `name` attribute (using an atomic counter), so multiple `RadioGroup`s on the same page don't share selection by accident.

---

## See also

- [Select](select.html) — the dropdown equivalent.
- [Checkbox](checkbox.html) — single boolean toggle.

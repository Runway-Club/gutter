---
title: Checkbox
parent: Widgets
nav_order: 22
---

# `Checkbox`
{: .no_toc }

A single boolean toggle rendered as the browser's native `<input type="checkbox">`, tinted with the theme's primary color via the CSS `accent-color` property.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Checkbox struct {
    Checked   bool
    Label     string
    Disabled  bool
    OnChanged func(bool)
    Name      string
}
```

| Field       | What it does                                                                          |
| ----------- | ------------------------------------------------------------------------------------- |
| `Checked`   | Source of truth. The widget syncs the DOM `checked` property on every rebuild.        |
| `Label`     | Optional text rendered next to the box, inside the same `<label>` so clicks toggle.   |
| `Disabled`  | Disables interaction.                                                                 |
| `OnChanged` | Fires with the new boolean on toggle.                                                 |
| `Name`      | Form field name.                                                                      |

---

## Basic usage

```go
widgets.Checkbox{
    Checked:   s.agreed,
    Label:     "I agree to the terms",
    OnChanged: func(v bool) { s.SetState(func() { s.agreed = v }) },
}
```

---

## Cross-tree state

For a checkbox whose value drives behavior elsewhere in the tree, pair with a `Notifier[bool]`:

```go
agreed := gutter.NewNotifier(false)

widgets.ObserverBuilder[bool]{
    Source: agreed,
    Builder: func(_ *gutter.BuildContext, v bool) gutter.Widget {
        return widgets.Checkbox{
            Checked:   v,
            Label:     "I agree",
            OnChanged: func(b bool) { agreed.Set(b) },
        }
    },
}
```

---

## See also

- [Switch](switch.html) — sliding-pill style toggle.
- [RadioGroup](radiogroup.html) — pick one of N.
- [Notifier + ObserverBuilder](observerbuilder.html).

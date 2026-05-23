---
title: Input
parent: Widgets
nav_order: 10
---

# `Input`
{: .no_toc }

A themed text field. Set `Error` to switch to the error-border variant. `OnChanged` fires for every keystroke.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Input struct {
    Value       string
    Placeholder string
    Error       bool
    OnChanged   func(string)
}
```

| Field         | What it does                                                                              |
| ------------- | ----------------------------------------------------------------------------------------- |
| `Value`       | Controlled value. Re-rendered on every rebuild; type to update via `OnChanged` + `SetState`. |
| `Placeholder` | Placeholder text shown when the input is empty.                                           |
| `Error`       | If true, the border uses the theme's `Input.BorderColorError`.                            |
| `OnChanged`   | Fires on the DOM `input` event with the new value. Wire it to a `SetState`.               |

Styling — background, foreground, border, rounded, padding, height, typography — comes from `theme.Components.Input`.

---

## The standard two-way binding

```go
type formState struct {
    gutter.StateObject
    name string
}

func (s *formState) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Input{
        Value:       s.name,
        Placeholder: "Your name",
        OnChanged:   func(v string) { s.SetState(func() { s.name = v }) },
    }
}
```

The element-tree reconciler updates the `<input>` element in place across rebuilds, so **focus is preserved** — the user can keep typing.

---

## Showing an error state

```go
widgets.Input{
    Value:     s.email,
    Error:     !isValidEmail(s.email),
    OnChanged: func(v string) { s.SetState(func() { s.email = v }) },
}
```

For a labeled error message below the input, compose:

```go
widgets.Column{
    Spacing: 4,
    Children: []gutter.Widget{
        widgets.Input{
            Value:     s.email,
            Error:     s.emailError != "",
            OnChanged: func(v string) { s.SetState(func() { s.email = v }) },
        },
        widgets.Caption{Text: s.emailError, Color: ctx.Theme.Colors.Critical},
    },
}
```

---

## Notes

- Only a single-line `<input type="text">` is provided. For passwords, multi-line, files, etc., drop to [`Styled`](styled.html):

  ```go
  widgets.Styled{
      Tag:   "input",
      Attrs: map[string]string{"type": "password", "value": s.password},
      // … your style + events …
  }
  ```

- `Input` is a `StatelessWidget`. The State that holds `Value` is your own widget's State, not `Input`'s — `Input` is purely declarative.

- For form submission, wire the parent widget's button `OnPressed` to whatever submit function you have. There's no `Form` widget today.

---

## See also

- [State Management](../state-management.html) — the `SetState` pattern Input expects.
- [Themes](../themes.html) — `Components.Input` and its `BorderColorError`.

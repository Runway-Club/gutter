---
title: Input
parent: Widgets
nav_order: 10
---

# `Input`
{: .no_toc }

A themed single-line text field, covering all 13 standard HTML input types via the `Type` field. Controlled — your `Value` field is the source of truth, `OnChanged` fires on every keystroke.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Input struct {
    Type        InputType
    Value       string
    Placeholder string
    Error       bool
    Disabled    bool
    ReadOnly    bool
    OnChanged   func(string)

    Min, Max, Step string
    Pattern        string
    AutoComplete   string
    Name           string
}
```

| Field          | What it does                                                                                 |
| -------------- | -------------------------------------------------------------------------------------------- |
| `Type`         | One of the [InputType constants](#input-types). Defaults to `InputText`.                     |
| `Value`        | Controlled value. Synced into the DOM via the `value` property on every rebuild.             |
| `Placeholder`  | Placeholder text shown when the input is empty.                                              |
| `Error`        | If true, the border uses `theme.Components.Input.BorderColorError`.                          |
| `Disabled`     | Disables interaction; the input is rendered with reduced opacity.                            |
| `ReadOnly`     | The user can focus and copy but not edit.                                                    |
| `OnChanged`    | Fires on the DOM `input` event with the new value. Wire it to a `SetState`.                  |
| `Min/Max/Step` | Numeric / date constraints. Forwarded to the matching HTML attribute. Step accepts `"any"`.  |
| `Pattern`      | Regex for client-side validation (text-like types).                                          |
| `AutoComplete` | Forwarded to `autocomplete` — e.g. `"email"`, `"current-password"`, `"off"`.                 |
| `Name`         | Form field name for native form submission.                                                  |

Styling — background, foreground, border, rounded, padding, height, typography — comes from `theme.Components.Input`. The `accent-color` CSS property pulls from `theme.Colors.Primary` so date and color pickers match the theme.

---

## Input types

```go
const (
    InputText          InputType = "text"
    InputPassword      InputType = "password"
    InputEmail         InputType = "email"
    InputNumber        InputType = "number"
    InputTel           InputType = "tel"
    InputURL           InputType = "url"
    InputSearch        InputType = "search"
    InputDate          InputType = "date"
    InputTime          InputType = "time"
    InputDateTimeLocal InputType = "datetime-local"
    InputMonth         InputType = "month"
    InputWeek          InputType = "week"
    InputColor         InputType = "color"
)
```

All variants share the theme's input styling — only the keyboard layout, validation, and platform widget (e.g. native date picker) change.

---

## The standard two-way binding

```go
type formState struct {
    gutter.StateObject
    name string
}

func (s *formState) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Input{
        Type:        widgets.InputText,
        Value:       s.name,
        Placeholder: "Your name",
        OnChanged:   func(v string) { s.SetState(func() { s.name = v }) },
    }
}
```

The element-tree reconciler updates the `<input>` in place across rebuilds, so **focus is preserved** — the user can keep typing. The `value` is set via the DOM property (not the attribute) and only when it actually differs from what the input already holds, so the caret stays where the user left it even when typing in the middle of a string.

---

## Showing an error state

```go
widgets.Input{
    Type:      widgets.InputEmail,
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
            Type:      widgets.InputEmail,
            Value:     s.email,
            Error:     s.emailError != "",
            OnChanged: func(v string) { s.SetState(func() { s.email = v }) },
        },
        widgets.Caption{Text: s.emailError, Color: ctx.Theme.Colors.Critical},
    },
}
```

---

## Constrained numeric input

```go
widgets.Input{
    Type:      widgets.InputNumber,
    Value:     s.quantity,
    Min:       "1",
    Max:       "99",
    Step:      "1",
    OnChanged: func(v string) { s.SetState(func() { s.quantity = v }) },
}
```

`Value` stays a string; convert with `strconv.Atoi` (or similar) when you need a number.

---

## Date and time pickers

```go
widgets.Input{Type: widgets.InputDate, Min: "2024-01-01"}
widgets.Input{Type: widgets.InputTime}
widgets.Input{Type: widgets.InputDateTimeLocal}
widgets.Input{Type: widgets.InputMonth}
widgets.Input{Type: widgets.InputWeek}
widgets.Input{Type: widgets.InputColor, Value: "#0066cc"}
```

The browser provides the picker UI for each; styling stays the theme's input box.

---

## Notes

- `Input` is a `StatefulWidget`-backed widget; the State that holds `Value` is *your* widget's State, not `Input`'s. `Input` itself is declarative.
- For multi-line text use [`TextArea`](textarea.html).
- For form submission, wire the parent widget's button `OnPressed` to your submit function. There's no `Form` widget today.

---

## See also

- [State Management](../state-management.html) — the `SetState` pattern.
- [TextArea](textarea.html) — multi-line counterpart.
- [Select](select.html), [Checkbox](checkbox.html), [Switch](switch.html), [Slider](slider.html), [RadioGroup](radiogroup.html) — the rest of the form-input family.
- [Themes](../themes.html) — `Components.Input` and its `BorderColorError`.

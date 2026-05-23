---
title: Styled
parent: Widgets
nav_order: 18
---

# `Styled`
{: .no_toc }

The escape hatch. Renders any HTML tag with arbitrary attributes, inline CSS, event handlers, children, and optional text content. Themed widgets build on top of `Styled`; app code rarely needs it directly.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Styled struct {
    Tag      string                       // defaults to "div"
    Text     string                       // textContent
    Attrs    map[string]string            // id, href, type, value, …
    Style    map[string]string            // inline CSS (camel- or kebab-case as in CSS)
    Events   map[string]func(gutter.Event) // listener functions, by DOM event name
    Children []gutter.Widget
}
```

| Field      | What it does                                                           |
| ---------- | ---------------------------------------------------------------------- |
| `Tag`      | The HTML tag to render. Defaults to `"div"`.                           |
| `Text`     | Sets `textContent`. Mutually exclusive with `Children` in practice.    |
| `Attrs`    | HTML attributes. Keys match the attribute name (`href`, `type`, …).    |
| `Style`    | Inline CSS. Keys are CSS property names (`background-color`, etc.).    |
| `Events`   | DOM event listeners. Keys are event names (`click`, `input`, `focus`). |
| `Children` | Sub-widgets, mounted recursively.                                      |

---

## Usage

### An external link

```go
widgets.Styled{
    Tag:   "a",
    Text:  "Read the docs",
    Attrs: map[string]string{"href": "https://example.com", "target": "_blank"},
    Style: map[string]string{
        "color":           ctx.Theme.Colors.Primary,
        "text-decoration": "underline",
    },
}
```

### A custom semantic element

```go
widgets.Styled{
    Tag: "section",
    Style: map[string]string{"max-width": "720px", "margin": "0 auto"},
    Children: []gutter.Widget{
        widgets.Heading{Level: widgets.H2, Text: "Article"},
        widgets.Body{Text: "…"},
    },
}
```

### A flex spacer

```go
widgets.Row{
    Children: []gutter.Widget{
        widgets.Text{Data: "Left"},
        widgets.Styled{Style: map[string]string{"flex": "1"}}, // spacer
        widgets.Text{Data: "Right"},
    },
}
```

### Listening for custom events

```go
widgets.Styled{
    Tag: "input",
    Attrs: map[string]string{"type": "checkbox"},
    Events: map[string]func(gutter.Event){
        "change": func(e gutter.Event) {
            // e.Value is the new value, when the DOM populates event.target.value
        },
    },
}
```

---

## Notes

- For inputs, `e.Value` is populated from `event.target.value`. For other event types it may be empty — read what you need via the underlying DOM if necessary.
- Empty `Style` / `Attrs` / `Events` maps are harmless; omit them if you don't need them.
- If you find yourself reaching for `Styled` repeatedly with the same pattern, wrap it in a small `StatelessWidget` to give the pattern a name.

---

## See also

- [`Container`](container.html) — when a styled `<div>` is what you need.
- [`Text`](text.html) — when a styled `<span>` is what you need.
- [Architecture](../architecture.html) — how `Host` and `Styled` relate.

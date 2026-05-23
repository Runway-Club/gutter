---
title: Text
parent: Widgets
nav_order: 17
---

# `Text`
{: .no_toc }

Renders a string as a `<span>` with an optional inline-CSS `TextStyle`. The low-level typography primitive that themed widgets like [`Heading`](heading.html) and [`Body`](body.html) build on top of.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Text struct {
    Data  string
    Style *TextStyle  // optional — nil means "no inline style"
}

type TextStyle struct {
    Color         string
    FontFamily    string
    FontSize      string
    FontWeight    string
    LineHeight    string
    LetterSpacing string
}
```

`TextStyle` is the inline-CSS shape for typography. All fields are CSS values; empty fields are omitted from the rendered style.

---

## Usage

```go
// Plain unstyled text — picks up cascaded font from <body>:
widgets.Text{Data: "Hello"}

// Styled:
widgets.Text{
    Data: "Important",
    Style: &widgets.TextStyle{
        Color:      "#d72d3c",
        FontWeight: "600",
        FontSize:   "16px",
    },
}
```

---

## When to use Text vs Heading vs Body

For 95% of app code, **don't use Text**. Use [`Heading`](heading.html), [`Body`](body.html), [`Caption`](caption.html), or [`Link`](link.html) — they pull typography from the active theme so your sizes, weights, and tracking are consistent.

Reach for `Text` only when:

- You're inside a custom widget that needs raw control over the inline CSS.
- You're rendering a label inside another widget (e.g. a custom Button child) and you want the parent's color/typography to cascade.
- You're writing tests or sketching layouts and don't want to depend on a theme.

---

## See also

- [`Heading`](heading.html), [`Body`](body.html), [`Caption`](caption.html), [`Link`](link.html) — themed typography.
- [`Styled`](styled.html) — for non-`<span>` raw elements.

---
title: Link
parent: Widgets
nav_order: 6
---

# `Link`
{: .no_toc }

Themed inline anchor. Renders an `<a>` styled with the theme's `Link` typography and `Primary` color, with an optional click handler.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Link struct {
    Text      string
    OnPressed func()
    Color     string
}
```

| Field       | What it does                                                                                      |
| ----------- | ------------------------------------------------------------------------------------------------- |
| `Text`      | The link text.                                                                                    |
| `OnPressed` | Click handler. If nil, the link is **non-interactive** — still styled, but no click is wired up.  |
| `Color`     | Optional override. Defaults to `ctx.Theme.Colors.Primary`.                                        |

The underlying tag is `<a href="javascript:void(0)">` — Gutter doesn't ship a router, so `Link` is a styled span-with-pointer-cursor, not a real navigation primitive. If you need an external link, use [`Styled`](styled.html) directly:

```go
widgets.Styled{
    Tag:   "a",
    Text:  "Read the docs",
    Attrs: map[string]string{"href": "https://example.com", "target": "_blank"},
    Style: map[string]string{"color": ctx.Theme.Colors.Primary},
}
```

---

## Usage

```go
widgets.Link{Text: "Learn more", OnPressed: func() { showLearnMore() }}

// Non-interactive (e.g. inside a breadcrumb that's not clickable):
widgets.Link{Text: "Current page"}

// Custom color:
widgets.Link{
    Text:      "Forgot password?",
    Color:     ctx.Theme.Colors.InkMuted,
    OnPressed: func() { goToReset() },
}
```

---

## See also

- [`Button`](button.html) — for primary calls to action, use `Button` not `Link`.
- [`Styled`](styled.html) — when you need an actual `<a href="…">`.

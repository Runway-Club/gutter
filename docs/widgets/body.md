---
title: Body
parent: Widgets
nav_order: 4
---

# `Body`
{: .no_toc }

Renders text in one of the theme's body roles. Toggle `Bold` for emphasis and `Small` for caption sizing.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Body struct {
    Text  string
    Bold  bool
    Small bool
    Color string
}
```

| Field   | What it does                                                          |
| ------- | --------------------------------------------------------------------- |
| `Text`  | The string to render.                                                 |
| `Bold`  | If true, use the strong variant of whichever role is selected.        |
| `Small` | If true, drop to caption sizing.                                      |
| `Color` | Optional override. Defaults to `ctx.Theme.Colors.Ink`.                |

The combinations map to four theme roles:

| `Bold` | `Small` | Reads from `Typography.*` |
| :----: | :-----: | ------------------------- |
| false  | false   | `Body`                    |
| true   | false   | `BodyStrong`              |
| false  | true    | `Caption`                 |
| true   | true    | `CaptionStrong`           |

---

## Usage

```go
widgets.Body{Text: "The quick brown fox jumps over the lazy dog."}
widgets.Body{Text: "Emphasized", Bold: true}
widgets.Body{Text: "Tiny", Small: true}
widgets.Body{Text: "Tiny + emphasized", Bold: true, Small: true}
```

For small text you can also use the shorthand [`Caption`](caption.html):

```go
widgets.Caption{Text: "Tap to learn more"}
// equivalent to
widgets.Body{Text: "Tap to learn more", Small: true}
```

---

## Colored body text

```go
// muted secondary text
widgets.Body{
    Text:  "Optional",
    Color: ctx.Theme.Colors.InkMuted,
}

// on a dark surface
widgets.Surface{
    Variant: widgets.SurfaceDark,
    Child: widgets.Body{
        Text:  "Welcome",
        Color: ctx.Theme.Colors.OnDark,
    },
}
```

---

## See also

- [`Caption`](caption.html) — shorthand for `Body{Small: true}`.
- [`Heading`](heading.html) — `H1`–`H6` headlines.
- [`Link`](link.html) — themed inline anchor.

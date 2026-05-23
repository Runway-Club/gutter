---
title: Heading
parent: Widgets
nav_order: 3
---

# `Heading`
{: .no_toc }

Renders text in one of the active theme's heading roles — `H1` (hero display) through `H6` (smallest section heading).
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Heading struct {
    Level HeadingLevel
    Text  string
    Color string
}
```

| Field   | What it does                                                                              |
| ------- | ----------------------------------------------------------------------------------------- |
| `Level` | One of `H1`, `H2`, `H3`, `H4`, `H5`, `H6`.                                                |
| `Text`  | The string to render.                                                                     |
| `Color` | Optional override. Defaults to `ctx.Theme.Colors.Ink`. Set to `Colors.OnDark` on dark surfaces. |

---

## Level mapping

| Level | Theme role                | Typical use                       |
| ----- | ------------------------- | --------------------------------- |
| `H1`  | `Typography.HeroDisplay`  | Top-of-page hero, marketing splash |
| `H2`  | `Typography.DisplayLarge` | Section title                     |
| `H3`  | `Typography.DisplayMedium`| Sub-section title                 |
| `H4`  | `Typography.HeadingLarge` | Card or column heading            |
| `H5`  | `Typography.HeadingMedium`| Feature-card title                |
| `H6`  | `Typography.HeadingSmall` | Smallest titled block             |

The actual font size, weight, line height, and letter spacing of each level come from the active theme. `Heading{Level: H1, Text: "Hello"}` is 56px / 600 / -0.28px under Apple, 64px / 500 / 0 under Meta.

---

## Usage

```go
widgets.Heading{Level: widgets.H1, Text: "One catalog. Two design systems."}

widgets.Heading{Level: widgets.H4, Text: "Feature"}
```

---

## On a dark surface

Heading defaults to the theme's `Ink` color, which is the wrong color on a dark background. Set `Color` explicitly:

```go
widgets.Surface{
    Variant: widgets.SurfaceDark,
    Child: widgets.Heading{
        Level: widgets.H1,
        Text:  "Hello",
        Color: ctx.Theme.Colors.OnDark,
    },
}
```

The pattern is reusable — anywhere you put `Heading` on `SurfaceDark`, `CardPromo`, or any tinted background, override `Color` from `ctx.Theme.Colors`.

---

## Notes

- `Heading` is a `StatelessWidget`. It builds to a `Text` (which is a `<span>`) with the theme-resolved CSS applied inline.
- `Heading` does **not** render `<h1>`–`<h6>` HTML tags. For semantic HTML in critical SEO paths, use `Styled{Tag: "h1", …}` directly.
- For body or caption-sized text, use [`Body`](body.html) or [`Caption`](caption.html) instead.

---

## See also

- [`Body`](body.html) — paragraph text in the theme's body roles.
- [Themes](../themes.html#typography--the-type-ladder) — the full typography ladder.

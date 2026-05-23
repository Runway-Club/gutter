---
title: Themes
nav_order: 5
---

# Themes
{: .no_toc }

How Gutter's theme system works, the three built-in presets, and how to write your own.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## The big idea

Gutter ships **pure-data themes**. The `themes` package imports nothing from the rest of the framework and renders nothing on its own — it's just tables of strings (CSS values). Every themed widget reads from those tables.

```text
┌──────────────────────┐    pointer    ┌──────────────────────┐
│  *themes.Theme       │ ◀──────────── │  BuildContext.Theme  │
│  - Colors            │               └──────────────────────┘
│  - Typography        │                          ▲
│  - Rounded           │                          │   reads
│  - Spacing           │                          │
│  - Components        │               ┌──────────────────────┐
└──────────────────────┘               │  widgets.Button{…}   │
                                       │  widgets.Card{…}     │
                                       │  widgets.Heading{…}  │
                                       └──────────────────────┘
```

App code says "I want a primary button"; the theme says what "primary" looks like in this design system.

```go
widgets.Button{Variant: widgets.ButtonPrimary, Label: "Buy"}

// → under themes.Apple  : Action Blue pill, 22px padding, 17px label
// → under themes.Meta   : true-black pill, 24px padding, 14px / 700 label
// → under themes.Neutral: medium-grey pill, system fonts
```

You never write CSS in app code.

---

## Choosing a theme

There are three places a theme can be chosen, in increasing precedence:

1. **Framework default** — `themes.Apple`, applied if you do nothing.
2. **`gutter.WithTheme(...)`** — passed to `RunApp`.
3. **`Scaffold{Theme: ...}`** — set on the app shell.

The Scaffold wins because it mutates `ctx.Theme` during its own Build, and the same `*BuildContext` is threaded through every descendant.

```go
// All three render an Apple-themed app:
gutter.RunApp(App{})
gutter.RunApp(App{}, gutter.WithTheme(themes.Apple))
gutter.RunApp(App{}) // where App.Build returns Scaffold{Theme: themes.Apple, …}

// Mix-and-match: WithTheme picks Meta, but Scaffold overrides to Apple.
gutter.RunApp(App{}, gutter.WithTheme(themes.Meta))
// where App.Build returns Scaffold{Theme: themes.Apple, …}
// → Apple wins.
```

Recommendation: **set the theme on `Scaffold`**. It's where everything else about the app shell lives — title, app bar, body, footer — so the theme belongs there too.

---

## Built-in presets

### `themes.Apple`

Extracted from [`theme_specs/APPLE_DESIGN.md`](https://github.com/Runway-Club/gutter/blob/main/theme_specs/APPLE_DESIGN.md). The look:

- Photography-first museum gallery.
- Single Action Blue (`#0066cc`) accent for every CTA.
- Pill primary buttons (`border-radius: 9999px`), 22px horizontal padding.
- 44px global-nav bar in pure black with 12px white labels.
- Parchment (`#f5f5f7`) + dark (`#272729`) alternating tile pattern.
- Apple-tight negative letter-spacing at display sizes.

The framework default. `gutter.RunApp(MyApp{})` with no options uses it.

### `themes.Meta`

Extracted from [`theme_specs/META_DESIGN.md`](https://github.com/Runway-Club/gutter/blob/main/theme_specs/META_DESIGN.md). The look:

- Hardware merchandiser.
- Dual-CTA pattern: black pill for marketing primary, cobalt (`#0064e0`) pill for commerce primary.
- 100px pill buttons, 32px rounded photographic cards.
- 64px white nav bar with a hairline-soft bottom border, 14px / 700 button-style label.

### `themes.Neutral`

Lexend-only, brand-agnostic fallback. Useful for unit tests or apps that don't want either of the brand looks. No fancy display fonts, no negative letter-spacing, no special CTA colors — just a sane neutral palette.

---

## Switching themes at runtime

Two cleanest patterns:

**1. Build-time switch** (compile two binaries, no runtime cost):

```go
var themeName = "apple" // overridden via -ldflags "-X main.themeName=meta"

func main() {
    theme := themes.Apple
    if themeName == "meta" {
        theme = themes.Meta
    }
    gutter.RunApp(MyApp{}, gutter.WithTheme(theme))
}
```

This is what `examples/showcase` does — same widget tree, two builds, two looks.

**2. Runtime switch** (one binary, theme on State):

```go
type themedApp struct{}
func (themedApp) CreateState() gutter.State { return &themedState{theme: themes.Apple} }

type themedState struct {
    gutter.StateObject
    theme *themes.Theme
}

func (s *themedState) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Scaffold{
        Theme: s.theme,
        AppBar: widgets.AppBar{
            Title: "My App",
            Actions: []gutter.Widget{
                widgets.Button{
                    Variant: widgets.ButtonGhost,
                    Label:   "Switch theme",
                    OnPressed: func() {
                        s.SetState(func() {
                            if s.theme == themes.Apple {
                                s.theme = themes.Meta
                            } else {
                                s.theme = themes.Apple
                            }
                        })
                    },
                },
            },
        },
        Body: /* … */,
    }
}
```

---

## Anatomy of a `*Theme`

Every preset implements the same shape:

```go
type Theme struct {
    Name       string
    Colors     Colors
    Typography Typography
    Rounded    Rounded
    Spacing    Spacing
    Components Components
}
```

### `Colors` — the semantic palette

Theme presets fill these by **role**, not by hex. Themed widgets reach for them by role.

| Field          | What it means                                                                  |
| -------------- | ------------------------------------------------------------------------------ |
| `Primary`      | Marketing primary CTA color. Apple Action Blue, Meta true-black.               |
| `OnPrimary`    | Text/icon color on top of `Primary`.                                           |
| `Accent`       | Secondary brand. Meta uses this for the cobalt commerce CTA.                   |
| `OnAccent`     | Text/icon color on top of `Accent`.                                            |
| `Canvas`       | Page background.                                                               |
| `CanvasAlt`    | Alternate light surface — parchment / soft cloud.                              |
| `SurfaceSoft`  | Tertiary soft surface.                                                         |
| `SurfaceDark`  | Dark tile / promo strip.                                                       |
| `OnDark`       | Text color on dark surfaces.                                                   |
| `Ink`          | Default body text on light.                                                    |
| `InkMuted`     | Secondary text — labels, captions.                                             |
| `InkSubtle`    | Tertiary text — fine print, disabled.                                          |
| `Hairline`     | Standard 1px border / divider color.                                           |
| `HairlineSoft` | Lighter hairline — subtle separators.                                          |
| `Success`      | Semantic success (positive).                                                   |
| `Warning`      | Semantic warning (caution).                                                    |
| `Critical`     | Semantic critical (destructive / error).                                       |

### `Typography` — the type ladder

A `Typography` field is a `TextSpec`:

```go
type TextSpec struct {
    FontFamily    string
    FontSize      string
    FontWeight    string
    LineHeight    string
    LetterSpacing string
}
```

Roles cover marketing display through legal fine print:

| Role            | Used by             |
| --------------- | ------------------- |
| `HeroDisplay`   | `Heading{Level: H1}`|
| `DisplayLarge`  | `Heading{Level: H2}`|
| `DisplayMedium` | `Heading{Level: H3}`|
| `HeadingLarge`  | `Heading{Level: H4}`|
| `HeadingMedium` | `Heading{Level: H5}`|
| `HeadingSmall`  | `Heading{Level: H6}`|
| `Lead`          | (your custom widgets)|
| `BodyStrong`    | `Body{Bold: true}`  |
| `Body`          | `Body{}`            |
| `Caption`       | `Body{Small: true}` / `Caption{}` |
| `CaptionStrong` | `Body{Small: true, Bold: true}`   |
| `Button`        | (read by button styles in `Components`) |
| `Link`          | `Link{}`            |
| `FinePrint`     | (your custom widgets)|

### `Rounded`, `Spacing`

CSS scales. Strings so themes can express both px and shorthand (`9999px` for pills, `50%` for circles).

### `Components` — pre-composed widget styles

`Components` is the connective tissue. Each themed widget reads exactly one field:

| Widget                                 | Reads from `Components.*`             |
| -------------------------------------- | ------------------------------------- |
| `Button{Variant: ButtonPrimary}`       | `ButtonPrimary`                       |
| `Button{Variant: ButtonSecondary}`     | `ButtonSecondary`                     |
| `Button{Variant: ButtonGhost}`         | `ButtonGhost`                         |
| `Button{Variant: ButtonAccent}`        | `ButtonAccent`                        |
| `Button{Variant: ButtonOnDark}`        | `ButtonOnDark`                        |
| `Card{Variant: CardFeature}`           | `CardFeature`                         |
| `Card{Variant: CardPromo}`             | `CardPromo`                           |
| `Card{Variant: CardPlain}`             | `CardPlain`                           |
| `Surface{Variant: SurfaceCanvas}`      | `SurfaceCanvas`                       |
| `Surface{Variant: SurfaceAlt}`         | `SurfaceAlt`                          |
| `Surface{Variant: SurfaceDark}`        | `SurfaceDark`                         |
| `Input{}`                              | `Input`                               |
| `Badge{Variant: BadgeNeutral}`         | `BadgeNeutral`                        |
| `Badge{Variant: BadgeSuccess}`         | `BadgeSuccess`                        |
| `Badge{Variant: BadgeWarning}`         | `BadgeWarning`                        |
| `Badge{Variant: BadgeCritical}`        | `BadgeCritical`                       |
| `AppBar{}`                             | `NavBar`                              |

---

## Reading the theme from app code

Most apps shouldn't have to. The themed widgets do it. But if you're writing a custom widget that needs to match the theme — say, a hero band whose text should use `OnDark` because it sits on `SurfaceDark` — read from `ctx.Theme.Colors`:

```go
func (h Hero) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Surface{
        Variant: widgets.SurfaceDark,
        Child: widgets.Heading{
            Level: widgets.H1,
            Text:  h.Title,
            Color: ctx.Theme.Colors.OnDark, // explicit override
        },
    }
}
```

`ctx.Theme` is never nil during normal mounting. (It defaults to `themes.Apple` if neither `WithTheme` nor `Scaffold.Theme` is set.)

---

## Writing your own theme

A theme is just a `*themes.Theme` value. Build one from scratch, or start by copying a preset and tweaking.

```go
package mytheme

import "github.com/Runway-Club/gutter/themes"

var MyTheme = &themes.Theme{
    Name: "MyTheme",
    Colors: themes.Colors{
        Primary:   "#7c3aed",  // violet
        OnPrimary: "#ffffff",
        // … fill every Colors field
    },
    Typography: themes.Typography{
        HeroDisplay: themes.TextSpec{
            FontFamily: "Lexend, system-ui, sans-serif",
            FontSize:   "56px",
            FontWeight: "600",
            LineHeight: "1.1",
        },
        // … fill every Typography field
    },
    Rounded: themes.Rounded{
        None: "0", Small: "4px", Medium: "8px", Large: "12px",
        XLarge: "16px", XXLarge: "24px", Pill: "9999px", Circle: "9999px",
    },
    Spacing: themes.Spacing{
        XXS: "4px", XS: "8px", SM: "12px", MD: "16px", LG: "24px",
        XL: "32px", XXL: "48px", XXXL: "64px", Section: "80px", Hero: "120px",
    },
    Components: themes.Components{
        ButtonPrimary: themes.ButtonStyle{
            Background: "#7c3aed", Foreground: "#ffffff",
            Rounded: "9999px", PaddingY: "10px", PaddingX: "20px",
            Typography: themes.TextSpec{FontSize: "15px", FontWeight: "600"},
        },
        // … every Components field a themed widget might read
    },
}

// Then:
gutter.RunApp(MyApp{}, gutter.WithTheme(MyTheme))
```

**The contract is "fill every field a themed widget reads from."** If you leave `Components.ButtonGhost` empty and then use `Button{Variant: ButtonGhost}`, you'll get an unstyled button — empty CSS values mean "don't write that property."

Easiest path: copy `themes/apple.go` to a new file, change the values, give the variable a new name. The structure is fixed, so a copy-and-tweak approach won't miss any fields.

---

## A word on fonts

Every built-in theme leads its font stack with [**Lexend**](https://www.lexend.com/), loaded from Google Fonts. The brand-specific fonts (SF Pro on Apple, Optimistic VF on Meta) remain as platform fallbacks behind it.

This is intentional: SF Pro is only available on Apple platforms, Optimistic VF isn't publicly distributed, and serving either yourself is a licensing question. Lexend is open-licensed and readable everywhere.

The `index.html` that `gutter new` scaffolds preloads Lexend via:

```html
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Lexend:wght@100..900&display=swap" rel="stylesheet">
```

If you write your own `index.html`, add those lines yourself or the font stack will fall through to system fonts.

---

## See also

- [`themes/theme.go`](https://github.com/Runway-Club/gutter/blob/main/themes/theme.go) — full type definitions.
- [`themes/apple.go`](https://github.com/Runway-Club/gutter/blob/main/themes/apple.go), [`themes/meta.go`](https://github.com/Runway-Club/gutter/blob/main/themes/meta.go), [`themes/neutral.go`](https://github.com/Runway-Club/gutter/blob/main/themes/neutral.go) — the presets.
- [`theme_specs/`](https://github.com/Runway-Club/gutter/tree/main/theme_specs) — the source-of-truth YAML+prose specs the Apple and Meta presets are extracted from.

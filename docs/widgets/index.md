---
title: Widgets
nav_order: 8
has_children: true
permalink: /widgets/
---

# Widgets

Gutter's standard catalog. Everything you need to build a real screen lives in the [`github.com/Runway-Club/gutter/widgets`](https://github.com/Runway-Club/gutter/blob/main/widgets) package.

Widgets fall into two loose groups:

- **Themed widgets** read styling from the active theme on `BuildContext`. App code never writes CSS — pick a variant, the theme picks the values.
- **Layout & primitive widgets** carry no theme dependency. Use them as building blocks for custom widgets or escape hatches when you need raw control.

## Catalog

| Widget                              | Theme-aware | Type             | Use for                                          |
| ----------------------------------- | :---------: | ---------------- | ------------------------------------------------ |
| [Scaffold](scaffold.html)           |     yes     | Stateless        | The app shell: title + theme + bar + body + footer. |
| [AppBar](appbar.html)               |     yes     | Stateless        | Top nav strip inside a Scaffold.                 |
| [Heading](heading.html)             |     yes     | Stateless        | Display + heading typography (`H1`–`H6`).        |
| [Body](body.html)                   |     yes     | Stateless        | Body text (`Bold`, `Small`).                     |
| [Caption](caption.html)             |     yes     | Stateless        | Shorthand for `Body{Small: true}`.               |
| [Link](link.html)                   |     yes     | Stateless        | Themed inline anchor.                            |
| [Button](button.html)               |     yes     | Stateless        | Primary, secondary, ghost, accent, on-dark.      |
| [Card](card.html)                   |     yes     | Stateless        | Feature, promo, plain.                           |
| [Surface](surface.html)             |     yes     | Stateless        | Full-bleed regions: canvas, alt, dark.           |
| [Input](input.html)                 |     yes     | Stateless        | Themed text field.                               |
| [Badge](badge.html)                 |     yes     | Stateless        | Status pill: neutral, success, warning, critical. |
| [Column / Row](column-row.html)     |     no      | Host (primitive) | Flex layouts.                                    |
| [Center](center.html)               |     no      | Host (primitive) | Center a single child in a full-size box.        |
| [Padding](padding.html)             |     no      | Host (primitive) | Wrap a child with `EdgeInsets`.                  |
| [SizedBox](sizedbox.html)           |     no      | Host (primitive) | Force fixed width / height.                      |
| [Container](container.html)         |     no      | Host (primitive) | Low-level styled `<div>` — raw colors/borders/radii. |
| [Text](text.html)                   |     no      | Host (primitive) | Raw `<span>` with explicit `TextStyle`.          |
| [Styled](styled.html)               |     no      | Host (primitive) | Escape hatch — any tag, arbitrary attrs/style/events. |
| [WithKey](withkey.html)             |     no      | Stateless wrapper | Add a reconciliation key to any child.           |

## When to reach for what

- **Start with `Scaffold`.** It's the canonical root. The whole catalog assumes there's a theme on the context, and `Scaffold` is where you set it.
- **Prefer themed widgets over primitives.** `Heading`, `Body`, `Button`, `Card`, `Surface`, `Input`, `Badge`, `Link` cover the vast majority of UI. They're consistent across themes.
- **Reach for layout primitives** when you need to compose: `Column`, `Row`, `Center`, `Padding`, `SizedBox`.
- **Drop to `Container`, `Text`, `Styled`** only when you need raw control — typically inside your own custom widgets, not directly in app code.

---
title: Widgets
nav_order: 8
has_children: true
permalink: /widgets/
---

# Widgets

Gutter's standard catalog. Everything you need to build a real screen lives in the [`github.com/Runway-Club/gutter/widgets`](https://github.com/Runway-Club/gutter/blob/main/widgets) package. Vendor-specific reusable widgets (e.g. Google sign-in) live under [`community/`](../community.html).

Widgets fall into a handful of loose groups:

- **App shell + themed** widgets read styling from the active theme on `BuildContext`. App code never writes CSS — pick a variant, the theme picks the values.
- **Form inputs** — single-line `Input` with 13 HTML types, multi-line `TextArea`, and the checkbox/switch/slider/select/radio family. Controlled inputs: the field is source of truth.
- **Layout & primitives** carry no theme dependency. Use them as building blocks for custom widgets or escape hatches when you need raw control.
- **Overlays** — modal/drawer/sheet, all driven by a `Listenable[bool]`.
- **Reactive / control flow** — observer pattern, async loading, time-driven animation, path-based routing.
- **Imperative** — escape hatches with direct DOM access via `Host.OnMount`.

## App shell + themed

| Widget                                     | Use for                                                       |
| ------------------------------------------ | ------------------------------------------------------------- |
| [Scaffold](scaffold.html)                  | App shell: title + theme + bar + body + footer (+ sticky bar). |
| [AppBar](appbar.html)                      | Top nav strip inside a Scaffold.                              |
| [Heading](heading.html)                    | Display + heading typography (`H1`–`H6`).                     |
| [Body](body.html)                          | Body text (`Bold`, `Small`).                                  |
| [Caption](caption.html)                    | Shorthand for `Body{Small: true}`.                            |
| [Link](link.html)                          | Themed inline anchor.                                         |
| [Button](button.html)                      | Primary, secondary, ghost, accent, on-dark.                   |
| [IconButton](iconbutton.html)              | Square Button rendering an Icon.                              |
| [Card](card.html)                          | Feature, promo, plain.                                        |
| [Surface](surface.html)                    | Full-bleed regions: canvas, alt, dark.                        |
| [Badge](badge.html)                        | Status pill: neutral, success, warning, critical.             |
| [Image](image.html)                        | `<img>` with `Asset` (via `gutter.AssetURL`) or `Src`.        |
| [Icon](icon.html)                          | Material Symbols glyph (outlined / rounded / sharp).          |
| [File](file.html)                          | Themed file picker — reads bytes via FileReader.              |

## Form inputs

| Widget                                     | Use for                                                       |
| ------------------------------------------ | ------------------------------------------------------------- |
| [Input](input.html)                        | Single-line text. 13 HTML types via `Type` (text, password, email, number, date, color, …). |
| [TextArea](textarea.html)                  | Multi-line text.                                              |
| [Checkbox](checkbox.html)                  | Single boolean toggle.                                        |
| [Switch](switch.html)                      | Sliding-pill toggle.                                          |
| [Slider](slider.html)                      | Range slider (`<input type="range">`).                        |
| [Select&lt;T&gt;](select.html)             | Generic dropdown with typed options.                          |
| [RadioGroup&lt;T&gt;](radiogroup.html)     | Generic radio button group.                                   |

## Layout & primitives

| Widget                                     | Use for                                                       |
| ------------------------------------------ | ------------------------------------------------------------- |
| [Column / Row](column-row.html)            | Flex layouts.                                                 |
| [Center](center.html)                      | Center a single child in a full-size box.                     |
| [Padding](padding.html)                    | Wrap a child with `EdgeInsets`.                               |
| [SizedBox](sizedbox.html)                  | Force fixed width / height.                                   |
| [Container](container.html)                | Low-level styled `<div>` — raw colors/borders/radii.          |
| [Text](text.html)                          | Raw `<span>` with explicit `TextStyle`.                       |
| [Styled](styled.html)                      | Escape hatch — any tag, arbitrary attrs/style/events.         |
| [Transform](transform.html)                | CSS translate/rotate/scale/skew wrapper.                      |
| [Draggable + DropTarget](dragdrop.html)    | Pointer-based drag-and-drop kit (kanban, sortables).          |
| [WithKey](withkey.html)                    | Add a reconciliation key to any child.                        |
| [List](list.html)                          | Eager scrollable flex container.                              |
| [ListBuilder](list.html#listbuilder)       | Virtualized list — recycles DOM as you scroll 10k+ rows.      |

## Overlays

| Widget                                     | Use for                                                       |
| ------------------------------------------ | ------------------------------------------------------------- |
| [Popup](popup.html)                        | Centered modal dialog with backdrop.                          |
| [Drawer](drawer.html)                      | Side panel that slides in from left/right.                    |
| [BottomSheet](bottomsheet.html)            | Panel that slides up from the bottom.                         |

## Reactive / control flow

| Widget                                     | Use for                                                       |
| ------------------------------------------ | ------------------------------------------------------------- |
| [ObserverBuilder&lt;T&gt;](observerbuilder.html) | Rebuild a subtree when a `Listenable[T]` fires.         |
| [AsyncBuilder&lt;T&gt;](asyncbuilder.html) | Run a `func(ctx) (T, error)` and rebuild with the snapshot.   |
| [AnimationController + AnimatedBuilder](animation.html) | Time-driven interpolation between two values.    |
| [Router + RouterView](router.html)         | Path-based routing with `:param` capture and browser history. |

## Imperative

| Widget                                     | Use for                                                       |
| ------------------------------------------ | ------------------------------------------------------------- |
| [Canvas](canvas.html)                      | Typed 2D painter — charts, sparklines, signature pads.        |
| [GestureDetector](gesturedetector.html)    | Wrap a child with pointer/key event handlers.                 |
| [Worker](worker.html)                      | Offload heavy work to a Web Worker via an inline Go task.     |

## When to reach for what

- **Start with `Scaffold`.** It's the canonical root. The whole catalog assumes there's a theme on the context, and `Scaffold` is where you set it.
- **Prefer themed widgets over primitives.** `Heading`, `Body`, `Button`, `Card`, `Surface`, `Input` and friends cover the vast majority of UI. They're consistent across themes.
- **Reach for layout primitives** when you need to compose: `Column`, `Row`, `Center`, `Padding`, `SizedBox`.
- **For long lists, prefer `ListBuilder`.** It only mounts the visible window and recycles DOM via positional reconciliation as you scroll.
- **Cross-tree state goes through `Notifier` + `ObserverBuilder`.** Drop `SetState` in the owner, observe from far descendants.
- **Drop to `Container`, `Text`, `Styled`** only when you need raw control — typically inside your own custom widgets, not directly in app code.

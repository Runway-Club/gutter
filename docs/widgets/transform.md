---
title: Transform
parent: Widgets
nav_order: 45
---

# `Transform`
{: .no_toc }

Applies a CSS `transform` to its child without changing the child's layout box. Use it for static positioning tweaks (nudging an icon a few pixels) or as the rendering surface for an [`AnimationController`](animation.html) driving movement, rotation, or scale.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Transform struct {
    Child gutter.Widget

    TranslateX, TranslateY float64 // CSS pixels
    Rotate                 float64 // degrees
    Scale                  float64 // uniform; zero means 1
    ScaleX, ScaleY         float64 // per-axis overrides
    SkewX, SkewY           float64 // degrees
    Origin                 string  // CSS transform-origin
    Transition             string  // optional CSS transition shorthand
}
```

All fields are additive — leave a field zero to skip that component of the transform. The zero-value `Transform{}` is the identity (no transform applied).

| Field                  | Maps to CSS                                       |
| ---------------------- | ------------------------------------------------- |
| `TranslateX/Y`         | `translate(x, y)`                                 |
| `Rotate`               | `rotate(deg)`                                     |
| `Scale` / `ScaleX/Y`   | `scale(x, y)`                                     |
| `SkewX/Y`              | `skew(x, y)`                                      |
| `Origin`               | `transform-origin`                                |
| `Transition`           | `transition` shorthand for animating the result   |

---

## Static transform

```go
widgets.Transform{
    TranslateY: -2,
    Child:      widgets.Icon{Name: "favorite"},
}
```

---

## Driven by an AnimationController

```go
widgets.AnimatedBuilder{
    Controller: ctrl,
    Builder: func(_ *gutter.BuildContext, t float64) gutter.Widget {
        return widgets.Transform{
            TranslateX: 240 * t,
            Rotate:     360 * t,
            Scale:      0.5 + 0.5*t,
            Child:      widgets.Container{Width: "64px", Height: "64px", Color: "#0066cc"},
        }
    },
}
```

---

## Notes

- The wrapper uses `display: inline-block` so transforms apply correctly even to inline content. Wrap with a Column/Row if you need a block-level container.
- `Scale` is a single value applied to both axes; if you set `ScaleX`/`ScaleY` they override per-axis.

---

## See also

- [AnimationController + AnimatedBuilder](animation.html) — the typical driver.

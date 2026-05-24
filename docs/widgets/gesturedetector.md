---
title: GestureDetector
parent: Widgets
nav_order: 61
---

# `GestureDetector`
{: .no_toc }

Wraps a child with DOM-level event listeners without changing the child's layout. Maps to a `display: contents` wrapper so the child's box model is preserved.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type GestureDetector struct {
    Child         gutter.Widget
    OnTap         func()
    OnDoubleTap   func()
    OnPointerDown func(gutter.Event)
    OnPointerMove func(gutter.Event)
    OnPointerUp   func(gutter.Event)
    OnKeyDown     func(gutter.Event)
    OnKeyUp       func(gutter.Event)
}
```

Each handler maps to one DOM event:

| Field           | DOM event       |
| --------------- | --------------- |
| `OnTap`         | `click`         |
| `OnDoubleTap`   | `dblclick`      |
| `OnPointerDown` | `pointerdown`   |
| `OnPointerMove` | `pointermove`   |
| `OnPointerUp`   | `pointerup`     |
| `OnKeyDown`     | `keydown`       |
| `OnKeyUp`       | `keyup`         |

The `Event` passed to pointer/mouse handlers carries:

- `X`, `Y` — viewport-relative coordinates
- `OffsetX`, `OffsetY` — coordinates in the target element's local space (useful for `Canvas`)

Keyboard handlers receive `Event.Key`.

---

## Basic usage

```go
widgets.GestureDetector{
    OnTap: func() { log.Println("tapped") },
    Child: widgets.Container{
        Width: "200px", Height: "100px",
        Color: "#0066cc",
        Child: widgets.Center{Child: widgets.Body{Text: "Click me"}},
    },
}
```

---

## Tracking the pointer

```go
type state struct {
    gutter.StateObject
    x, y float64
}

func (s *state) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.GestureDetector{
        OnPointerMove: func(e gutter.Event) {
            s.SetState(func() { s.x = e.OffsetX; s.y = e.OffsetY })
        },
        Child: widgets.Container{
            Width: "100%", Height: "200px",
            Color: "#0066cc20",
            Child: widgets.Center{
                Child: widgets.Caption{
                    Text: fmt.Sprintf("(%.0f, %.0f)", s.x, s.y),
                },
            },
        },
    }
}
```

---

## Notes

- The wrapper uses `display: contents` so it doesn't add an extra box to the layout — your child renders exactly as it would without GestureDetector.
- For touch devices, pointer events cover the common cases. If you need swipe gestures, layer the handler yourself on top of `OnPointerDown`/`Move`/`Up`.

---

## See also

- [Canvas](canvas.html) — typical pair for interactive drawing.

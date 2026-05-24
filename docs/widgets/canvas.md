---
title: Canvas
parent: Widgets
nav_order: 60
---

# `Canvas`
{: .no_toc }

A typed wrapper over the browser's `<canvas>` 2D rendering context. Use it for charts, sparklines, signature pads, games — anything imperative that doesn't map cleanly to declarative DOM.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signature

```go
type Canvas struct {
    Width      float64
    Height     float64
    Background string
    Paint      func(p *CanvasPainter)
}
```

| Field        | What it does                                                                |
| ------------ | --------------------------------------------------------------------------- |
| `Width/Height` | CSS-pixel dimensions of the canvas.                                       |
| `Background` | Optional CSS background color.                                             |
| `Paint`      | Painter callback invoked after every mount and every update.                |

`CanvasPainter` mirrors the common methods of `CanvasRenderingContext2D`: `Clear`, `FillStyle/StrokeStyle`, `FillRect/StrokeRect`, paths (`BeginPath`, `MoveTo`, `LineTo`, `Arc`, `Rect`, `ClosePath`, `Fill`, `Stroke`), text (`Font`, `FillText`, `StrokeText`, `TextAlign`, `TextBaseline`), and transforms (`Save`, `Restore`, `Translate`, `Rotate`, `Scale`).

The backing-store is scaled by `devicePixelRatio` so strokes stay crisp on retina displays. Your painter always works in logical (CSS) pixels via `p.Size()`.

---

## Basic usage

```go
widgets.Canvas{
    Width:  480,
    Height: 160,
    Paint: func(p *widgets.CanvasPainter) {
        w, h := p.Size()
        p.Clear()
        p.FillStyle("#0066cc")
        p.FillRect(0, 0, w, h)
        p.FillStyle("#fff")
        p.Font("16px Lexend")
        p.TextBaseline("top")
        p.FillText("Hello canvas", 16, 16)
    },
}
```

---

## A simple bar chart

```go
bars := []float64{0.3, 0.7, 0.5, 0.9, 0.4, 0.65}
widgets.Canvas{
    Width: 480, Height: 160, Background: "#f5f5f7",
    Paint: func(p *widgets.CanvasPainter) {
        w, h := p.Size()
        p.Clear()
        barW := w / float64(len(bars)*2)
        for i, v := range bars {
            x := float64(i)*barW*2 + barW*0.5
            barH := v * (h - 32)
            p.FillStyle("#0066cc")
            p.FillRect(x, h-barH-8, barW, barH)
        }
    },
}
```

---

## Notes

- `Paint` runs on every reconcile. Cheap painters can just redraw; expensive ones should compare against memoized state before doing the work.
- The widget lives in `canvas_wasm.go` / `canvas_stub.go` so it builds on the host too — `CanvasPainter` methods are no-ops off WASM.

---

## See also

- [GestureDetector](gesturedetector.html) — wrap a Canvas with pointer events for interactive painting.

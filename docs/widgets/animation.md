---
title: AnimationController
parent: Widgets
nav_order: 72
---

# `AnimationController` + `AnimatedBuilder`
{: .no_toc }

Drive a `float64` between two values over time, using a `Curve` (linear, ease-in, ease-out, ease-in-out). The controller is a `Listenable[float64]` — pair it with `AnimatedBuilder` (a typed alias of `ObserverBuilder[float64]`) for a clean call site, or use [`Transform`](transform.html) directly to animate movement.
{: .fs-6 .fw-300 }

1. TOC
{:toc}

---

## Signatures

```go
type Curve func(t float64) float64

func CurveLinear(t float64) float64
func CurveEaseIn(t float64) float64
func CurveEaseOut(t float64) float64
func CurveEaseInOut(t float64) float64

type AnimationController struct {
    Duration time.Duration
    Curve    Curve
    Lower    float64
    Upper    float64
    // … unexported
}

func NewAnimationController(duration time.Duration) *AnimationController

// Methods on *AnimationController:
//   Value() float64
//   Listen(fn func(float64)) (cancel func())
//   Forward(), Reverse(), Reset(), Stop()
```

`Forward` animates `Lower → Upper`; `Reverse` does the opposite. Calling either while a previous run is in flight cancels it.

```go
type AnimatedBuilder struct {
    Controller *AnimationController
    Builder    func(ctx *gutter.BuildContext, value float64) gutter.Widget
}
```

`AnimatedBuilder` is just a semantic alias of `ObserverBuilder[float64]` — both work, but `AnimatedBuilder` reads more clearly when the source is a controller.

---

## Basic usage

```go
type myState struct {
    gutter.StateObject
    anim *widgets.AnimationController
}

func (s *myState) InitState() {
    s.anim = widgets.NewAnimationController(900 * time.Millisecond)
    s.anim.Curve = widgets.CurveEaseInOut
}

func (s *myState) Dispose() {
    if s.anim != nil { s.anim.Stop() }
}

func (s *myState) Build(ctx *gutter.BuildContext) gutter.Widget {
    return widgets.Column{
        Spacing: 12,
        Children: []gutter.Widget{
            widgets.Button{Label: "Forward", OnPressed: func() { s.anim.Forward() }},
            widgets.Button{Label: "Reverse", OnPressed: func() { s.anim.Reverse() }},
            widgets.AnimatedBuilder{
                Controller: s.anim,
                Builder: func(_ *gutter.BuildContext, t float64) gutter.Widget {
                    return widgets.Transform{
                        TranslateX: 240 * t,
                        Rotate:     360 * t,
                        Scale:      0.5 + 0.5*t,
                        Child:      widgets.Container{Width: "64px", Height: "64px", Color: "#0066cc"},
                    }
                },
            },
        },
    }
}
```

---

## Notes

- The controller ticks at 60 Hz from a goroutine via `time.NewTicker`. No CSS animation framework; works cross-platform.
- Always call `Stop()` (or `Reset()`) in `Dispose()` so the goroutine releases.
- For custom curves, write your own `Curve` function: `func(t float64) float64` mapping `[0,1] → [0,1]`.

---

## See also

- [Transform](transform.html) — the typical rendering surface.
- [ObserverBuilder](observerbuilder.html) — same listening machinery.

package widgets

import (
	"math"
	"sync"
	"time"

	"github.com/Runway-Club/gutter"
)

// Curve maps an elapsed-fraction t in [0,1] to an output fraction in the same
// range. Built-in curves cover linear motion and the three standard ease
// variants; for anything fancier supply your own closure.
type Curve func(t float64) float64

// CurveLinear is the identity curve.
func CurveLinear(t float64) float64 { return t }

// CurveEaseIn accelerates from rest (t^2).
func CurveEaseIn(t float64) float64 { return t * t }

// CurveEaseOut decelerates to rest (1 - (1-t)^2).
func CurveEaseOut(t float64) float64 { return 1 - (1-t)*(1-t) }

// CurveEaseInOut blends ease-in and ease-out.
func CurveEaseInOut(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - math.Pow(-2*t+2, 2)/2
}

// AnimationController drives a float64 between Lower and Upper over Duration
// using a Curve. It implements gutter.Listenable[float64] so observers (most
// commonly an AnimatedBuilder, but any ObserverBuilder[float64] works) rebuild
// on each tick.
//
// The controller does not own a goroutine until Forward/Reverse is called and
// it stops the goroutine when the animation finishes or Stop is called. It is
// safe to call Forward/Reverse repeatedly; each call cancels the previous run.
//
// A typical usage pattern is to construct the controller in a State's
// InitState, drive it from event handlers, and call Stop in Dispose so the
// goroutine is released when the widget unmounts.
type AnimationController struct {
	Duration time.Duration
	Curve    Curve
	Lower    float64
	Upper    float64

	value  *gutter.Notifier[float64]
	mu     sync.Mutex
	cancel chan struct{}
}

// NewAnimationController returns a controller with Lower=0, Upper=1, and a
// linear curve, seeded at Lower. Adjust the exported fields after construction
// if you want a different range or curve.
func NewAnimationController(duration time.Duration) *AnimationController {
	return &AnimationController{
		Duration: duration,
		Curve:    CurveLinear,
		Lower:    0,
		Upper:    1,
		value:    gutter.NewNotifier(0.0),
	}
}

// Value returns the current animated value. Implements gutter.Listenable.
func (c *AnimationController) Value() float64 { return c.value.Value() }

// Listen subscribes fn to value changes. Implements gutter.Listenable.
func (c *AnimationController) Listen(fn func(float64)) func() { return c.value.Listen(fn) }

// Forward animates from Lower to Upper.
func (c *AnimationController) Forward() { c.run(c.Lower, c.Upper) }

// Reverse animates from Upper to Lower.
func (c *AnimationController) Reverse() { c.run(c.Upper, c.Lower) }

// Reset jumps to Lower without animating and stops any in-flight run.
func (c *AnimationController) Reset() {
	c.Stop()
	c.value.Set(c.Lower)
}

// Stop cancels any running animation. The current value is left as-is.
func (c *AnimationController) Stop() {
	c.mu.Lock()
	if c.cancel != nil {
		close(c.cancel)
		c.cancel = nil
	}
	c.mu.Unlock()
}

func (c *AnimationController) run(from, to float64) {
	c.Stop()
	c.mu.Lock()
	cancel := make(chan struct{})
	c.cancel = cancel
	c.mu.Unlock()

	curve := c.Curve
	if curve == nil {
		curve = CurveLinear
	}
	dur := c.Duration
	if dur <= 0 {
		c.value.Set(to)
		return
	}
	c.value.Set(from)
	start := time.Now()
	go func() {
		// 60Hz tick is fine for CSS-transform-driven motion; the browser
		// is the actual rasterizer so we're just feeding it intermediate
		// numbers.
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-cancel:
				return
			case now := <-ticker.C:
				t := float64(now.Sub(start)) / float64(dur)
				if t >= 1 {
					c.value.Set(to)
					return
				}
				c.value.Set(from + (to-from)*curve(t))
			}
		}
	}()
}

// AnimatedBuilder rebuilds Builder on every tick of Controller. It is a
// semantic alias for ObserverBuilder[float64] — both work the same way, but
// AnimatedBuilder reads more clearly at the call site when the source is a
// time-driven controller.
type AnimatedBuilder struct {
	Controller *AnimationController
	Builder    func(ctx *gutter.BuildContext, value float64) gutter.Widget
}

func (a AnimatedBuilder) Build(ctx *gutter.BuildContext) gutter.Widget {
	if a.Controller == nil || a.Builder == nil {
		return nil
	}
	return ObserverBuilder[float64]{
		Source:  a.Controller,
		Builder: a.Builder,
	}
}

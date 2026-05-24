package widgets

import (
	"fmt"
	"strings"

	"github.com/Runway-Club/gutter"
)

// Transform applies a CSS transform to its child without changing the child's
// layout box. Use it for static positioning tweaks (e.g. nudging an icon by a
// few pixels) or as the rendering surface for an [AnimationController] driving
// movement, rotation, or scale.
//
// All fields are additive — leave a field zero to skip that component of the
// transform. Scale defaults to 1 when both ScaleX and ScaleY are zero, so the
// zero-value Transform is the identity.
//
// Pair with AnimatedBuilder for time-driven motion:
//
//	widgets.AnimatedBuilder{
//	    Controller: ctrl,
//	    Builder: func(_ *gutter.BuildContext, t float64) gutter.Widget {
//	        return widgets.Transform{
//	            Scale: 0.5 + 0.5*t,
//	            Child: widgets.Heading{Level: widgets.H1, Text: "Hello"},
//	        }
//	    },
//	}
type Transform struct {
	Child gutter.Widget

	// Translation in CSS pixels.
	TranslateX, TranslateY float64

	// Rotation in degrees.
	Rotate float64

	// Uniform scale. When non-zero, applied to both axes. ScaleX/ScaleY
	// override per-axis. The zero value (all three zero) is treated as 1.
	Scale          float64
	ScaleX, ScaleY float64

	// Skew in degrees.
	SkewX, SkewY float64

	// transform-origin (CSS); empty falls back to "50% 50%".
	Origin string

	// Optional CSS transition shorthand applied alongside the transform —
	// useful when you want a snappy click effect without an
	// AnimationController, e.g. "transform 0.15s ease-out".
	Transition string
}

func (t Transform) Build(ctx *gutter.BuildContext) gutter.Widget {
	style := map[string]string{}
	if v := t.transformString(); v != "" {
		style["transform"] = v
	}
	if t.Origin != "" {
		style["transform-origin"] = t.Origin
	}
	if t.Transition != "" {
		style["transition"] = t.Transition
	}
	// display:inline-block keeps transforms working on what would otherwise
	// be inline content; users wrapping a block child can override via Styled.
	style["display"] = "inline-block"

	children := []gutter.Widget(nil)
	if t.Child != nil {
		children = []gutter.Widget{t.Child}
	}
	return Styled{Style: style, Children: children}
}

func (t Transform) transformString() string {
	var parts []string
	if t.TranslateX != 0 || t.TranslateY != 0 {
		parts = append(parts, fmt.Sprintf("translate(%gpx, %gpx)", t.TranslateX, t.TranslateY))
	}
	if t.Rotate != 0 {
		parts = append(parts, fmt.Sprintf("rotate(%gdeg)", t.Rotate))
	}
	sx, sy := t.ScaleX, t.ScaleY
	if t.Scale != 0 {
		if sx == 0 {
			sx = t.Scale
		}
		if sy == 0 {
			sy = t.Scale
		}
	}
	if sx != 0 || sy != 0 {
		if sx == 0 {
			sx = 1
		}
		if sy == 0 {
			sy = 1
		}
		parts = append(parts, fmt.Sprintf("scale(%g, %g)", sx, sy))
	}
	if t.SkewX != 0 || t.SkewY != 0 {
		parts = append(parts, fmt.Sprintf("skew(%gdeg, %gdeg)", t.SkewX, t.SkewY))
	}
	return strings.Join(parts, " ")
}

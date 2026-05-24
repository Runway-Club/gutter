package widgets

import (
	"fmt"

	"github.com/Runway-Club/gutter"
)

// Canvas renders an HTML <canvas> element and invokes Paint after every
// mount and update with a typed painter. Use it for charts, sparklines,
// games, signature pads, or anything imperative that doesn't map cleanly
// to the declarative DOM.
//
// Width and Height are CSS-pixel dimensions. The framework also adjusts the
// canvas's backing-store size to match devicePixelRatio so strokes stay
// crisp on retina displays — Paint always receives logical (CSS) pixel
// coordinates via CanvasPainter.Size.
//
// Paint runs on every rebuild of the surrounding subtree. Cheap painters
// can simply redraw; expensive ones should compare against memoized state
// before doing the work.
type Canvas struct {
	Width      float64
	Height     float64
	Background string
	Paint      func(p *CanvasPainter)
}

func (c Canvas) Host() *gutter.Host {
	style := map[string]string{"display": "block"}
	if c.Width > 0 {
		style["width"] = fmt.Sprintf("%gpx", c.Width)
	}
	if c.Height > 0 {
		style["height"] = fmt.Sprintf("%gpx", c.Height)
	}
	if c.Background != "" {
		style["background"] = c.Background
	}
	painter := c.Paint
	w, h := c.Width, c.Height
	return &gutter.Host{
		Tag:   "canvas",
		Style: style,
		OnMount: func(node any) {
			paintCanvas(node, w, h, painter)
		},
	}
}

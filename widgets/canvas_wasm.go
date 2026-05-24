//go:build js && wasm

package widgets

import (
	"fmt"
	"syscall/js"
)

// CanvasPainter is the typed wrapper over a 2D rendering context handed
// to Canvas.Paint. It mirrors the most commonly used portions of the
// CanvasRenderingContext2D API — enough to draw shapes, paths, and text
// without the caller ever touching syscall/js directly.
type CanvasPainter struct {
	ctx js.Value
	w   float64
	h   float64
}

// Size returns the logical (CSS-pixel) drawing area. The backing-store
// resolution is higher on retina displays, but the painter operates in
// logical coordinates because the context is pre-scaled by devicePixelRatio.
func (p *CanvasPainter) Size() (float64, float64) { return p.w, p.h }

// Clear erases the entire drawable area.
func (p *CanvasPainter) Clear() { p.ctx.Call("clearRect", 0, 0, p.w, p.h) }

// ClearRect erases the given rectangle.
func (p *CanvasPainter) ClearRect(x, y, w, h float64) {
	p.ctx.Call("clearRect", x, y, w, h)
}

// FillStyle sets the current fill color, gradient, or pattern (as a CSS string).
func (p *CanvasPainter) FillStyle(v string) { p.ctx.Set("fillStyle", v) }

// StrokeStyle sets the current stroke color.
func (p *CanvasPainter) StrokeStyle(v string) { p.ctx.Set("strokeStyle", v) }

// LineWidth sets the stroke width in logical pixels.
func (p *CanvasPainter) LineWidth(v float64) { p.ctx.Set("lineWidth", v) }

// LineCap sets the line ending style: "butt", "round", or "square".
func (p *CanvasPainter) LineCap(v string) { p.ctx.Set("lineCap", v) }

// LineJoin sets the line corner style: "miter", "round", or "bevel".
func (p *CanvasPainter) LineJoin(v string) { p.ctx.Set("lineJoin", v) }

// Font sets the current font spec (e.g. "16px Lexend").
func (p *CanvasPainter) Font(v string) { p.ctx.Set("font", v) }

// TextAlign sets horizontal text alignment: "start", "end", "left", "right", "center".
func (p *CanvasPainter) TextAlign(v string) { p.ctx.Set("textAlign", v) }

// TextBaseline sets vertical baseline: "top", "middle", "bottom", "alphabetic", etc.
func (p *CanvasPainter) TextBaseline(v string) { p.ctx.Set("textBaseline", v) }

// FillRect paints a filled rectangle in the current fill style.
func (p *CanvasPainter) FillRect(x, y, w, h float64) { p.ctx.Call("fillRect", x, y, w, h) }

// StrokeRect outlines a rectangle in the current stroke style.
func (p *CanvasPainter) StrokeRect(x, y, w, h float64) { p.ctx.Call("strokeRect", x, y, w, h) }

// BeginPath starts a new path; previous subpaths are discarded.
func (p *CanvasPainter) BeginPath() { p.ctx.Call("beginPath") }

// ClosePath connects the current subpath back to its starting point.
func (p *CanvasPainter) ClosePath() { p.ctx.Call("closePath") }

// MoveTo lifts the pen and moves it to (x, y).
func (p *CanvasPainter) MoveTo(x, y float64) { p.ctx.Call("moveTo", x, y) }

// LineTo draws a straight line from the current position to (x, y).
func (p *CanvasPainter) LineTo(x, y float64) { p.ctx.Call("lineTo", x, y) }

// Rect adds a rectangle subpath to the current path.
func (p *CanvasPainter) Rect(x, y, w, h float64) { p.ctx.Call("rect", x, y, w, h) }

// Arc traces a circular arc centered at (x, y), radius r, between start and end (radians).
func (p *CanvasPainter) Arc(x, y, r, start, end float64) {
	p.ctx.Call("arc", x, y, r, start, end)
}

// Fill fills the current path with the current fill style.
func (p *CanvasPainter) Fill() { p.ctx.Call("fill") }

// Stroke strokes the current path with the current stroke style.
func (p *CanvasPainter) Stroke() { p.ctx.Call("stroke") }

// FillText paints text at (x, y) using the current font and fill style.
func (p *CanvasPainter) FillText(text string, x, y float64) {
	p.ctx.Call("fillText", text, x, y)
}

// StrokeText outlines text at (x, y).
func (p *CanvasPainter) StrokeText(text string, x, y float64) {
	p.ctx.Call("strokeText", text, x, y)
}

// Save pushes the current drawing state onto the stack.
func (p *CanvasPainter) Save() { p.ctx.Call("save") }

// Restore pops the most recent drawing state off the stack.
func (p *CanvasPainter) Restore() { p.ctx.Call("restore") }

// Translate adds a translation to the current transform.
func (p *CanvasPainter) Translate(x, y float64) { p.ctx.Call("translate", x, y) }

// Rotate adds a rotation (radians) to the current transform.
func (p *CanvasPainter) Rotate(rad float64) { p.ctx.Call("rotate", rad) }

// Scale scales the current transform by (sx, sy).
func (p *CanvasPainter) Scale(sx, sy float64) { p.ctx.Call("scale", sx, sy) }

// paintCanvas resolves the backing-store sizing and invokes the user's
// painter. It runs on every mount and every update — Canvas reuses the
// same DOM element across updates, so resetting the transform each time
// prevents accumulated scale from prior frames.
func paintCanvas(node any, width, height float64, paint func(*CanvasPainter)) {
	n, ok := node.(js.Value)
	if !ok {
		return
	}
	dpr := js.Global().Get("devicePixelRatio").Float()
	if dpr == 0 {
		dpr = 1
	}
	if width <= 0 {
		width = n.Get("clientWidth").Float()
	}
	if height <= 0 {
		height = n.Get("clientHeight").Float()
	}
	n.Set("width", int(width*dpr))
	n.Set("height", int(height*dpr))
	style := n.Get("style")
	style.Set("width", fmt.Sprintf("%gpx", width))
	style.Set("height", fmt.Sprintf("%gpx", height))
	if paint == nil {
		return
	}
	ctx := n.Call("getContext", "2d")
	ctx.Call("setTransform", dpr, 0, 0, dpr, 0, 0)
	paint(&CanvasPainter{ctx: ctx, w: width, h: height})
}

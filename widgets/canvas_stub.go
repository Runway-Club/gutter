//go:build !js || !wasm

package widgets

// CanvasPainter has no-op methods on the host platform. The Canvas widget
// only paints when running under GOOS=js GOARCH=wasm; on host builds the
// painter exists so user code that references it (e.g. an inline
// `func(p *widgets.CanvasPainter)` literal) still compiles for editor
// tooling and `go vet`.
type CanvasPainter struct{}

func (p *CanvasPainter) Size() (float64, float64)             { return 0, 0 }
func (p *CanvasPainter) Clear()                               {}
func (p *CanvasPainter) ClearRect(x, y, w, h float64)         {}
func (p *CanvasPainter) FillStyle(v string)                   {}
func (p *CanvasPainter) StrokeStyle(v string)                 {}
func (p *CanvasPainter) LineWidth(v float64)                  {}
func (p *CanvasPainter) LineCap(v string)                     {}
func (p *CanvasPainter) LineJoin(v string)                    {}
func (p *CanvasPainter) Font(v string)                        {}
func (p *CanvasPainter) TextAlign(v string)                   {}
func (p *CanvasPainter) TextBaseline(v string)                {}
func (p *CanvasPainter) FillRect(x, y, w, h float64)          {}
func (p *CanvasPainter) StrokeRect(x, y, w, h float64)        {}
func (p *CanvasPainter) BeginPath()                           {}
func (p *CanvasPainter) ClosePath()                           {}
func (p *CanvasPainter) MoveTo(x, y float64)                  {}
func (p *CanvasPainter) LineTo(x, y float64)                  {}
func (p *CanvasPainter) Rect(x, y, w, h float64)              {}
func (p *CanvasPainter) Arc(x, y, r, start, end float64)      {}
func (p *CanvasPainter) Fill()                                {}
func (p *CanvasPainter) Stroke()                              {}
func (p *CanvasPainter) FillText(text string, x, y float64)   {}
func (p *CanvasPainter) StrokeText(text string, x, y float64) {}
func (p *CanvasPainter) Save()                                {}
func (p *CanvasPainter) Restore()                             {}
func (p *CanvasPainter) Translate(x, y float64)               {}
func (p *CanvasPainter) Rotate(rad float64)                   {}
func (p *CanvasPainter) Scale(sx, sy float64)                 {}

func paintCanvas(node any, width, height float64, paint func(*CanvasPainter)) {}

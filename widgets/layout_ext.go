package widgets

import (
	"fmt"

	"github.com/Runway-Club/gutter"
)

// px formats a pixel length, returning "" for a non-positive value so callers
// can omit the property entirely.
func px(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%gpx", v)
}

// one wraps a single optional child in a *Host with the given tag/style.
func box(style map[string]string, child gutter.Widget) *gutter.Host {
	h := &gutter.Host{Tag: "div", Style: style}
	if child != nil {
		h.Children = []gutter.Widget{child}
	}
	return h
}

// =========== flex children ===========

// Expanded makes its child fill the free space along the main axis of the
// enclosing Row or Column, the way Flutter's Expanded does. Flex weights the
// share when several Expanded siblings compete; the zero value means 1.
//
// Use this instead of writing Styled{Style:{"flex":"1"}} by hand — it also
// sets min-width/min-height:0 so a long child can shrink below its content
// size rather than forcing the row to overflow (the classic flexbox gotcha).
type Expanded struct {
	Child gutter.Widget
	Flex  int
}

func (e Expanded) Host() *gutter.Host {
	flex := e.Flex
	if flex <= 0 {
		flex = 1
	}
	return box(map[string]string{
		"flex":       fmt.Sprintf("%d 1 0%%", flex),
		"min-width":  "0",
		"min-height": "0",
	}, e.Child)
}

// Flexible is like Expanded but keeps the child's natural size as its basis: it
// may grow into free space but won't force itself to zero. Flex weights how
// growth is shared; the zero value means 1.
type Flexible struct {
	Child gutter.Widget
	Flex  int
}

func (f Flexible) Host() *gutter.Host {
	flex := f.Flex
	if flex <= 0 {
		flex = 1
	}
	return box(map[string]string{"flex": fmt.Sprintf("%d 1 auto", flex)}, f.Child)
}

// Spacer is flexible empty space inside a Row or Column — handy for pushing
// siblings apart (e.g. a title on the left, actions on the right). Flex weights
// it against other Spacers/Expanded; the zero value means 1.
type Spacer struct {
	Flex int
}

func (s Spacer) Host() *gutter.Host {
	flex := s.Flex
	if flex <= 0 {
		flex = 1
	}
	return &gutter.Host{Tag: "div", Style: map[string]string{"flex": fmt.Sprintf("%d 1 0%%", flex)}}
}

// =========== stack / positioned ===========

// Stack layers its children in a positioning context. Plain children flow
// normally and size the stack; wrap any child in Positioned to pin it with
// absolute offsets over the others (badges, overlays, corner ribbons). Width
// and Height size the stack explicitly when no in-flow child does.
type Stack struct {
	Children []gutter.Widget
	Width    string
	Height   string
}

func (s Stack) Host() *gutter.Host {
	style := map[string]string{"position": "relative"}
	if s.Width != "" {
		style["width"] = s.Width
	}
	if s.Height != "" {
		style["height"] = s.Height
	}
	return &gutter.Host{Tag: "div", Style: style, Children: s.Children}
}

// Positioned pins its child with absolute offsets inside the nearest Stack.
// Offsets and size are CSS lengths ("0", "8px", "50%"); empty means unset.
// Fill is a shortcut that stretches the child to all four edges (inset:0).
type Positioned struct {
	Child                    gutter.Widget
	Top, Right, Bottom, Left string
	Width, Height            string
	Fill                     bool
}

func (p Positioned) Host() *gutter.Host {
	style := map[string]string{"position": "absolute"}
	if p.Fill {
		style["inset"] = "0"
	}
	setIf(style, "top", p.Top)
	setIf(style, "right", p.Right)
	setIf(style, "bottom", p.Bottom)
	setIf(style, "left", p.Left)
	setIf(style, "width", p.Width)
	setIf(style, "height", p.Height)
	return box(style, p.Child)
}

// =========== grid ===========

// Grid arranges children with CSS Grid. Pick exactly one column strategy:
//
//   - Columns: a fixed number of equal-width columns (repeat(N, 1fr)).
//   - MinColumnWidth: a responsive track that fits as many columns of at least
//     this width as the row allows (repeat(auto-fill, minmax(MIN, 1fr))) — this
//     is the no-media-query way to get a grid that reflows on its own.
//   - Template: a raw grid-template-columns string for full control.
//
// Template wins over MinColumnWidth, which wins over Columns. Gap sets both
// axes; RowGap/ColumnGap override per axis.
type Grid struct {
	Children       []gutter.Widget
	Columns        int
	MinColumnWidth string
	Template       string
	Gap            float64
	RowGap         float64
	ColumnGap      float64
	AlignItems     string
	JustifyItems   string
}

func (g Grid) Host() *gutter.Host {
	style := map[string]string{"display": "grid"}
	switch {
	case g.Template != "":
		style["grid-template-columns"] = g.Template
	case g.MinColumnWidth != "":
		style["grid-template-columns"] = fmt.Sprintf("repeat(auto-fill, minmax(%s, 1fr))", g.MinColumnWidth)
	case g.Columns > 0:
		style["grid-template-columns"] = fmt.Sprintf("repeat(%d, 1fr)", g.Columns)
	}
	if g.Gap > 0 {
		style["gap"] = px(g.Gap)
	}
	if g.RowGap > 0 {
		style["row-gap"] = px(g.RowGap)
	}
	if g.ColumnGap > 0 {
		style["column-gap"] = px(g.ColumnGap)
	}
	setIf(style, "align-items", g.AlignItems)
	setIf(style, "justify-items", g.JustifyItems)
	return &gutter.Host{Tag: "div", Style: style, Children: g.Children}
}

// =========== wrap ===========

// Wrap lays children out along the main axis like a Row/Column but wraps to a
// new line when they run out of room — chips, tags, responsive card strips.
// Spacing is the gap between items on a line; RunSpacing is the gap between
// lines. Direction defaults to horizontal.
type Wrap struct {
	Children       []gutter.Widget
	Direction      string // "row" (default) or "column"
	Spacing        float64
	RunSpacing     float64
	Alignment      string // justify-content
	CrossAlignment string // align-items
}

func (w Wrap) Host() *gutter.Host {
	dir := w.Direction
	if dir == "" {
		dir = "row"
	}
	style := map[string]string{
		"display":        "flex",
		"flex-wrap":      "wrap",
		"flex-direction": dir,
	}
	if w.RunSpacing > 0 || w.Spacing > 0 {
		// gap: <row-gap> <column-gap>. For a row, run spacing is vertical
		// (row-gap) and item spacing is horizontal (column-gap).
		row, col := w.RunSpacing, w.Spacing
		if dir == "column" {
			row, col = w.Spacing, w.RunSpacing
		}
		style["gap"] = fmt.Sprintf("%gpx %gpx", row, col)
	}
	setIf(style, "justify-content", w.Alignment)
	setIf(style, "align-items", w.CrossAlignment)
	return &gutter.Host{Tag: "div", Style: style, Children: w.Children}
}

// =========== align ===========

// Alignment positions a child within a box. Use the named presets
// (AlignCenter, AlignTopRight, …) rather than building one by hand.
type Alignment struct {
	Justify string // horizontal: justify-content
	Align   string // vertical: align-items
}

// Alignment presets covering the nine anchor points of a box.
var (
	AlignTopLeft      = Alignment{CrossAxisStart, CrossAxisStart}
	AlignTopCenter    = Alignment{MainAxisCenter, CrossAxisStart}
	AlignTopRight     = Alignment{MainAxisEnd, CrossAxisStart}
	AlignCenterLeft   = Alignment{CrossAxisStart, CrossAxisCenter}
	AlignCenter       = Alignment{MainAxisCenter, CrossAxisCenter}
	AlignCenterRight  = Alignment{MainAxisEnd, CrossAxisCenter}
	AlignBottomLeft   = Alignment{CrossAxisStart, CrossAxisEnd}
	AlignBottomCenter = Alignment{MainAxisCenter, CrossAxisEnd}
	AlignBottomRight  = Alignment{MainAxisEnd, CrossAxisEnd}
)

// Align positions its child at one of the nine anchor points of a full-size
// box. Center is the special case Align{Alignment: AlignCenter} covers — Center
// stays as its own widget for the common path.
type Align struct {
	Alignment Alignment
	Child     gutter.Widget
}

func (a Align) Host() *gutter.Host {
	just := a.Alignment.Justify
	if just == "" {
		just = MainAxisCenter
	}
	align := a.Alignment.Align
	if align == "" {
		align = CrossAxisCenter
	}
	return box(map[string]string{
		"display":         "flex",
		"justify-content": just,
		"align-items":     align,
		"width":           "100%",
		"height":          "100%",
	}, a.Child)
}

// =========== aspect ratio / constraints ===========

// AspectRatio forces its child into a fixed width:height proportion (Ratio is
// width divided by height: 16/9 for video, 1 for a square). By default the box
// takes the available width and derives its height; set Width to constrain it.
type AspectRatio struct {
	Ratio float64
	Width string
	Child gutter.Widget
}

func (a AspectRatio) Host() *gutter.Host {
	ratio := a.Ratio
	if ratio <= 0 {
		ratio = 1
	}
	style := map[string]string{
		"aspect-ratio": fmt.Sprintf("%g", ratio),
		"overflow":     "hidden",
	}
	if a.Width != "" {
		style["width"] = a.Width
	} else {
		style["width"] = "100%"
	}
	return box(style, a.Child)
}

// ConstrainedBox clamps its child's size with CSS min/max constraints. The
// most common use is a readable content column: ConstrainedBox{MaxWidth:
// "720px"} inside a Center. All fields are CSS lengths; empty means unset.
type ConstrainedBox struct {
	MinWidth, MaxWidth   string
	MinHeight, MaxHeight string
	Child                gutter.Widget
}

func (c ConstrainedBox) Host() *gutter.Host {
	style := map[string]string{}
	setIf(style, "min-width", c.MinWidth)
	setIf(style, "max-width", c.MaxWidth)
	setIf(style, "min-height", c.MinHeight)
	setIf(style, "max-height", c.MaxHeight)
	return box(style, c.Child)
}

package widgets

import "github.com/Runway-Club/gutter"

// GestureDetector wraps Child with DOM-level event listeners without
// changing the layout of the wrapped widget. It does not paint anything of
// its own; it merely attaches handlers to a transparent wrapper element.
//
// Each handler maps to one DOM event:
//
//	OnTap         -> "click"
//	OnDoubleTap   -> "dblclick"
//	OnPointerDown -> "pointerdown"
//	OnPointerMove -> "pointermove"
//	OnPointerUp   -> "pointerup"
//	OnKeyDown     -> "keydown"
//	OnKeyUp       -> "keyup"
//
// The Event passed to pointer/mouse handlers carries viewport-relative
// X/Y coordinates; keyboard handlers carry the Key string. Tap-style
// handlers (OnTap, OnDoubleTap) ignore the event payload since they only
// signal occurrence.
//
// The wrapper uses display:contents so it has no box of its own and the
// child's box model is unchanged. Pointer events bubble from the child up
// to the wrapper, which is where the listeners live.
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

func (g GestureDetector) Build(ctx *gutter.BuildContext) gutter.Widget {
	events := map[string]func(gutter.Event){}
	if g.OnTap != nil {
		f := g.OnTap
		events["click"] = func(gutter.Event) { f() }
	}
	if g.OnDoubleTap != nil {
		f := g.OnDoubleTap
		events["dblclick"] = func(gutter.Event) { f() }
	}
	if g.OnPointerDown != nil {
		events["pointerdown"] = g.OnPointerDown
	}
	if g.OnPointerMove != nil {
		events["pointermove"] = g.OnPointerMove
	}
	if g.OnPointerUp != nil {
		events["pointerup"] = g.OnPointerUp
	}
	if g.OnKeyDown != nil {
		events["keydown"] = g.OnKeyDown
	}
	if g.OnKeyUp != nil {
		events["keyup"] = g.OnKeyUp
	}
	children := []gutter.Widget(nil)
	if g.Child != nil {
		children = []gutter.Widget{g.Child}
	}
	return Styled{
		Style:    map[string]string{"display": "contents"},
		Events:   events,
		Children: children,
	}
}

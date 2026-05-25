package widgets

import (
	"github.com/Runway-Club/gutter"
)

// DrawerSide picks which edge the drawer slides in from.
type DrawerSide int

const (
	// DrawerLeft slides in from the left edge.
	DrawerLeft DrawerSide = iota
	// DrawerRight slides in from the right edge.
	DrawerRight
)

// Drawer is a side panel that slides in from the left or right edge of the
// viewport. Like [Popup], visibility is observed from a [gutter.Listenable]
// of bool so the open/close state lives in app code.
//
//	open := gutter.NewNotifier(false)
//	widgets.Drawer{
//	    Open: open,
//	    Side: widgets.DrawerLeft,
//	    OnDismiss: func() { open.Set(false) },
//	    Child: widgets.Column{Children: navItems},
//	}
type Drawer struct {
	// Open is observed for visibility transitions.
	Open gutter.Listenable[bool]

	// Child is rendered inside the drawer panel.
	Child gutter.Widget

	// Side picks the edge. Defaults to DrawerLeft.
	Side DrawerSide

	// Width is the panel width. Defaults to "min(80vw, 320px)".
	Width string

	// OnDismiss is invoked when the backdrop is clicked. Leave nil for a
	// non-dismissible drawer.
	OnDismiss func()

	// ZIndex is the base CSS z-index for the overlay layer. Defaults to
	// "1000".
	ZIndex string
}

func (d Drawer) Build(ctx *gutter.BuildContext) gutter.Widget {
	if d.Open == nil {
		return Styled{}
	}
	captured := d
	return ObserverBuilder[bool]{
		Source: d.Open,
		Builder: func(ctx *gutter.BuildContext, isOpen bool) gutter.Widget {
			return drawerRender(ctx, captured, isOpen)
		},
	}
}

func drawerRender(ctx *gutter.BuildContext, d Drawer, isOpen bool) gutter.Widget {
	t := activeTheme(ctx)
	z := fallback(d.ZIndex, "1000")
	width := fallback(d.Width, "min(80vw, 320px)")

	backdrop := overlayBackdrop(isOpen, z, d.OnDismiss)

	panelStyle := map[string]string{
		"position":   "fixed",
		"top":        "0",
		"bottom":     "0",
		"width":      width,
		"background": fallback(t.Colors.Canvas, "#ffffff"),
		"color":      fallback(t.Colors.Ink, "#000000"),
		"padding":    fallback(t.Spacing.LG, "24px"),
		"z-index":    z,
		"transition": "transform 0.25s ease-out",
		"box-shadow": "0 0 32px rgba(0,0,0,0.18)",
		"box-sizing": "border-box",
		"overflow-y": "auto",
	}
	closedTransform := "translateX(-100%)"
	if d.Side == DrawerRight {
		panelStyle["right"] = "0"
		closedTransform = "translateX(100%)"
	} else {
		panelStyle["left"] = "0"
	}
	if isOpen {
		panelStyle["transform"] = "translateX(0)"
	} else {
		panelStyle["transform"] = closedTransform
	}

	panelChildren := []gutter.Widget(nil)
	if d.Child != nil {
		panelChildren = []gutter.Widget{d.Child}
	}
	panel := Styled{Attrs: dialogAttrs(isOpen), Style: panelStyle, Children: panelChildren}

	return gutter.Portal{Child: Styled{
		Style:    map[string]string{"display": "contents"},
		Children: []gutter.Widget{backdrop, panel},
	}}
}

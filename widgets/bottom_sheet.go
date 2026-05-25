package widgets

import (
	"github.com/Runway-Club/gutter"
)

// BottomSheet is a panel that slides up from the bottom edge of the viewport.
// Visibility is observed from a [gutter.Listenable] of bool, mirroring [Popup]
// and [Drawer].
//
//	open := gutter.NewNotifier(false)
//	widgets.BottomSheet{
//	    Open: open,
//	    OnDismiss: func() { open.Set(false) },
//	    Child: widgets.Column{Children: actionItems},
//	}
type BottomSheet struct {
	// Open is observed for visibility transitions.
	Open gutter.Listenable[bool]

	// Child is rendered inside the sheet.
	Child gutter.Widget

	// Height caps the sheet's height. Defaults to "min(60vh, 480px)". The
	// sheet always grows from the bottom; reducing this leaves more of the
	// underlying app visible behind the backdrop.
	Height string

	// OnDismiss is invoked when the backdrop is clicked. Leave nil for a
	// non-dismissible sheet.
	OnDismiss func()

	// ZIndex is the base CSS z-index for the overlay layer. Defaults to
	// "1000".
	ZIndex string
}

func (b BottomSheet) Build(ctx *gutter.BuildContext) gutter.Widget {
	if b.Open == nil {
		return Styled{}
	}
	captured := b
	return ObserverBuilder[bool]{
		Source: b.Open,
		Builder: func(ctx *gutter.BuildContext, isOpen bool) gutter.Widget {
			return bottomSheetRender(ctx, captured, isOpen)
		},
	}
}

func bottomSheetRender(ctx *gutter.BuildContext, b BottomSheet, isOpen bool) gutter.Widget {
	t := activeTheme(ctx)
	z := fallback(b.ZIndex, "1000")
	height := fallback(b.Height, "min(60vh, 480px)")

	backdrop := overlayBackdrop(isOpen, z, b.OnDismiss)

	radius := fallback(t.Rounded.Large, "12px")
	sheetStyle := map[string]string{
		"position":                "fixed",
		"left":                    "0",
		"right":                   "0",
		"bottom":                  "0",
		"max-height":              height,
		"background":              fallback(t.Colors.Canvas, "#ffffff"),
		"color":                   fallback(t.Colors.Ink, "#000000"),
		"padding":                 fallback(t.Spacing.LG, "24px"),
		"border-top-left-radius":  radius,
		"border-top-right-radius": radius,
		"z-index":                 z,
		"transition":              "transform 0.25s ease-out",
		"box-shadow":              "0 -8px 32px rgba(0,0,0,0.18)",
		"box-sizing":              "border-box",
		"overflow-y":              "auto",
	}
	if isOpen {
		sheetStyle["transform"] = "translateY(0)"
	} else {
		sheetStyle["transform"] = "translateY(100%)"
	}

	sheetChildren := []gutter.Widget(nil)
	if b.Child != nil {
		sheetChildren = []gutter.Widget{b.Child}
	}
	sheet := Styled{Style: sheetStyle, Children: sheetChildren}

	return Styled{
		Style:    map[string]string{"display": "contents"},
		Children: []gutter.Widget{backdrop, sheet},
	}
}

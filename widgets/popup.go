package widgets

import (
	"github.com/Runway-Club/gutter"
)

// Popup is a centered modal dialog with a dim backdrop. Visibility is driven
// by a [gutter.Listenable] of bool — typically a [gutter.Notifier] the app
// holds and flips from a button — so the open/close state lives in app code
// and the popup is just a rendering of it.
//
//	open := gutter.NewNotifier(false)
//	widgets.Popup{
//	    Open: open,
//	    OnDismiss: func() { open.Set(false) },
//	    Child: widgets.Card{Child: widgets.Body{Text: "Hello"}},
//	}
//
// Popup is always mounted; the open/closed states differ only in CSS so the
// fade/scale transition runs in both directions. Click on the backdrop fires
// OnDismiss; clicks on the sheet itself bubble to a sibling overlay
// container and never reach the backdrop handler.
type Popup struct {
	// Open is observed for visibility transitions.
	Open gutter.Listenable[bool]

	// Child is rendered inside the popup sheet.
	Child gutter.Widget

	// OnDismiss is invoked when the backdrop is clicked. Leave nil for a
	// non-dismissible popup (the caller must close it programmatically).
	OnDismiss func()

	// MaxWidth caps the sheet width. Defaults to "min(90vw, 480px)".
	MaxWidth string

	// ZIndex is the base CSS z-index for the overlay layer. Defaults to
	// "1000". The sheet inherits the same value and stacks on top of the
	// backdrop by DOM order.
	ZIndex string
}

func (p Popup) Build(ctx *gutter.BuildContext) gutter.Widget {
	if p.Open == nil {
		return Styled{}
	}
	captured := p
	return ObserverBuilder[bool]{
		Source: p.Open,
		Builder: func(ctx *gutter.BuildContext, isOpen bool) gutter.Widget {
			return popupRender(ctx, captured, isOpen)
		},
	}
}

func popupRender(ctx *gutter.BuildContext, p Popup, isOpen bool) gutter.Widget {
	t := activeTheme(ctx)
	z := fallback(p.ZIndex, "1000")

	backdrop := overlayBackdrop(isOpen, z, p.OnDismiss)

	sheetStyle := map[string]string{
		"position":      "fixed",
		"top":           "50%",
		"left":          "50%",
		"background":    fallback(t.Colors.Canvas, "#ffffff"),
		"color":         fallback(t.Colors.Ink, "#000000"),
		"border-radius": fallback(t.Rounded.Large, "12px"),
		"padding":       fallback(t.Spacing.LG, "24px"),
		"max-width":     fallback(p.MaxWidth, "min(90vw, 480px)"),
		"z-index":       z,
		"transition":    "transform 0.2s ease-out, opacity 0.2s ease-out",
		"box-shadow":    "0 24px 60px rgba(0,0,0,0.25)",
		"box-sizing":    "border-box",
	}
	if isOpen {
		sheetStyle["opacity"] = "1"
		sheetStyle["pointer-events"] = "auto"
		sheetStyle["transform"] = "translate(-50%, -50%) scale(1)"
	} else {
		sheetStyle["opacity"] = "0"
		sheetStyle["pointer-events"] = "none"
		sheetStyle["transform"] = "translate(-50%, -50%) scale(0.95)"
	}

	sheetChildren := []gutter.Widget(nil)
	if p.Child != nil {
		sheetChildren = []gutter.Widget{p.Child}
	}
	sheet := Styled{Attrs: dialogAttrs(isOpen), Style: sheetStyle, Children: sheetChildren}

	// Teleport into the body-level portal root so the fixed backdrop/sheet aren't
	// trapped by an ancestor's transform/overflow/stacking context.
	return gutter.Portal{Child: Styled{
		Style:    map[string]string{"display": "contents"},
		Children: []gutter.Widget{backdrop, sheet},
	}}
}

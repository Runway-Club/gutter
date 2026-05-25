package widgets

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

// activeTheme returns the theme on ctx, falling back to the framework
// default (Apple) so widgets don't panic when used outside a normal RunApp
// (e.g. in unit tests that don't construct a BuildContext).
func activeTheme(ctx *gutter.BuildContext) *themes.Theme {
	if ctx != nil && ctx.Theme != nil {
		return ctx.Theme
	}
	return themes.Apple
}

// applySpec writes a TextSpec into a CSS style map. Empty fields are
// skipped — the rendered style only contains values the theme set.
func applySpec(style map[string]string, spec themes.TextSpec) {
	if spec.FontFamily != "" {
		style["font-family"] = spec.FontFamily
	}
	if spec.FontSize != "" {
		style["font-size"] = spec.FontSize
	}
	if spec.FontWeight != "" {
		style["font-weight"] = spec.FontWeight
	}
	if spec.LineHeight != "" {
		style["line-height"] = spec.LineHeight
	}
	if spec.LetterSpacing != "" {
		style["letter-spacing"] = spec.LetterSpacing
	}
}

func fallback(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

// dialogAttrs returns the ARIA attributes for a modal overlay sheet (Popup,
// Drawer, BottomSheet). role=dialog + aria-modal lets screen readers treat it
// as a modal; aria-hidden hides the always-mounted-but-closed sheet from the
// accessibility tree so it isn't reachable while invisible.
func dialogAttrs(open bool) map[string]string {
	a := map[string]string{"role": "dialog", "aria-modal": "true"}
	if !open {
		a["aria-hidden"] = "true"
	}
	return a
}

// propSyncHost is a HostWidget that exposes OnMount and OnUnmount alongside
// the usual tag/attrs/style/events bundle. It's used by input-style widgets
// (Checkbox, Switch, Slider, Select, RadioGroup) that need to imperatively
// sync DOM properties (checked, value) after every reconcile — applyAttrs
// only writes attributes, which for form elements just sets the default,
// not the live property the browser actually renders.
type propSyncHost struct {
	tag       string
	attrs     map[string]string
	style     map[string]string
	events    map[string]func(gutter.Event)
	children  []gutter.Widget
	text      string
	onMount   func(node any)
	onUnmount func(node any)
}

func (p propSyncHost) Host() *gutter.Host {
	return &gutter.Host{
		Tag:       p.tag,
		Text:      p.text,
		Attrs:     p.attrs,
		Style:     p.style,
		Events:    p.events,
		Children:  p.children,
		OnMount:   p.onMount,
		OnUnmount: p.onUnmount,
	}
}

// overlayBackdrop builds the dim full-viewport scrim used by Popup, Drawer,
// and BottomSheet. When open is false the scrim is transparent and ignores
// pointer events so clicks fall through to the underlying app. onDismiss may
// be nil for a non-dismissible overlay.
func overlayBackdrop(open bool, zIndex string, onDismiss func()) Styled {
	style := map[string]string{
		"position":   "fixed",
		"inset":      "0",
		"background": "rgba(0,0,0,0.45)",
		"z-index":    zIndex,
		"transition": "opacity 0.2s ease-out",
	}
	if open {
		style["opacity"] = "1"
		style["pointer-events"] = "auto"
	} else {
		style["opacity"] = "0"
		style["pointer-events"] = "none"
	}
	events := map[string]func(gutter.Event){}
	if onDismiss != nil {
		dismiss := onDismiss
		events["click"] = func(gutter.Event) { dismiss() }
	}
	return Styled{Style: style, Events: events}
}

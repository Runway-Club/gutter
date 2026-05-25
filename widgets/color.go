package widgets

import (
	"strings"

	"github.com/Runway-Club/gutter/themes"
)

// Color tokens name a role in the active theme's palette instead of a literal
// CSS color. Pass one anywhere a themed widget takes a color string (e.g.
// Container.Color, Container.BorderColor) and the widget resolves it against
// ctx.Theme.Colors at build time — so app code references intent ("ink",
// "primary") rather than hard-coding hex, and a theme swap recolors everything.
//
// These are ordinary strings, so a raw CSS color ("#fff", "rgb(...)",
// "tomato") still works in the same field; only values carrying the "theme:"
// sentinel are looked up. resolveColor leaves everything else untouched.
const (
	ColorPrimary      = "theme:primary"
	ColorOnPrimary    = "theme:on-primary"
	ColorAccent       = "theme:accent"
	ColorOnAccent     = "theme:on-accent"
	ColorCanvas       = "theme:canvas"
	ColorCanvasAlt    = "theme:canvas-alt"
	ColorSurfaceSoft  = "theme:surface-soft"
	ColorSurfaceDark  = "theme:surface-dark"
	ColorOnDark       = "theme:on-dark"
	ColorInk          = "theme:ink"
	ColorInkMuted     = "theme:ink-muted"
	ColorInkSubtle    = "theme:ink-subtle"
	ColorHairline     = "theme:hairline"
	ColorHairlineSoft = "theme:hairline-soft"
	ColorSuccess      = "theme:success"
	ColorWarning      = "theme:warning"
	ColorCritical     = "theme:critical"
)

// resolveColor maps a color value to a concrete CSS color. A value carrying
// the "theme:" sentinel is looked up in t.Colors; anything else (a raw CSS
// color, or "") is returned unchanged. An unknown token resolves to "" so it
// is simply omitted from the style rather than emitting a broken value.
func resolveColor(t *themes.Theme, v string) string {
	if v == "" || !strings.HasPrefix(v, "theme:") {
		return v
	}
	c := t.Colors
	switch v {
	case ColorPrimary:
		return c.Primary
	case ColorOnPrimary:
		return c.OnPrimary
	case ColorAccent:
		return c.Accent
	case ColorOnAccent:
		return c.OnAccent
	case ColorCanvas:
		return c.Canvas
	case ColorCanvasAlt:
		return c.CanvasAlt
	case ColorSurfaceSoft:
		return c.SurfaceSoft
	case ColorSurfaceDark:
		return c.SurfaceDark
	case ColorOnDark:
		return c.OnDark
	case ColorInk:
		return c.Ink
	case ColorInkMuted:
		return c.InkMuted
	case ColorInkSubtle:
		return c.InkSubtle
	case ColorHairline:
		return c.Hairline
	case ColorHairlineSoft:
		return c.HairlineSoft
	case ColorSuccess:
		return c.Success
	case ColorWarning:
		return c.Warning
	case ColorCritical:
		return c.Critical
	}
	return ""
}

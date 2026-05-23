package widgets

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

// CardVariant selects which entry in theme.Components.Card* to use.
type CardVariant int

const (
	// CardFeature is the bordered light card used for product/feature tiles.
	CardFeature CardVariant = iota
	// CardPromo is the dark promo surface (Meta's promo strip, Apple's dark
	// tile).
	CardPromo
	// CardPlain is a minimally-decorated rounded surface.
	CardPlain
)

// Card is a themed bordered/filled box with one child. Padding defaults to
// the theme's card padding; pass an explicit value to override.
type Card struct {
	Variant CardVariant
	Padding string
	Child   gutter.Widget
}

func (c Card) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := cardStyleFor(t, c.Variant)
	css := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"border-radius":    style.Rounded,
		"padding":          fallback(c.Padding, style.Padding),
		"box-sizing":       "border-box",
	}
	if style.BorderColor != "" && style.BorderWidth != "" {
		css["border"] = style.BorderWidth + " solid " + style.BorderColor
	}
	w := Styled{Style: css}
	if c.Child != nil {
		w.Children = []gutter.Widget{c.Child}
	}
	return w
}

func cardStyleFor(t *themes.Theme, variant CardVariant) themes.CardStyle {
	switch variant {
	case CardPromo:
		return t.Components.CardPromo
	case CardPlain:
		return t.Components.CardPlain
	default:
		return t.Components.CardFeature
	}
}

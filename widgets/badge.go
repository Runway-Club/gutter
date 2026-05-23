package widgets

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

// BadgeVariant selects one of the theme's status pill styles.
type BadgeVariant int

const (
	BadgeNeutral BadgeVariant = iota
	BadgeSuccess
	BadgeWarning
	BadgeCritical
)

// Badge is a small status pill used for "In stock", "Limited time",
// "Out of stock", etc.
type Badge struct {
	Variant BadgeVariant
	Text    string
}

func (b Badge) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := badgeStyleFor(t, b.Variant)
	css := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"border-radius":    style.Rounded,
		"padding":          style.Padding,
		"display":          "inline-flex",
		"align-items":      "center",
	}
	applySpec(css, style.Typography)
	return Styled{Tag: "span", Text: b.Text, Style: css}
}

func badgeStyleFor(t *themes.Theme, variant BadgeVariant) themes.BadgeStyle {
	switch variant {
	case BadgeSuccess:
		return t.Components.BadgeSuccess
	case BadgeWarning:
		return t.Components.BadgeWarning
	case BadgeCritical:
		return t.Components.BadgeCritical
	default:
		return t.Components.BadgeNeutral
	}
}

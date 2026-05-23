package widgets

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

// ButtonVariant selects which entry in theme.Components is consulted for the
// button's styling. The same variant produces visually different buttons
// across themes — that's the point.
type ButtonVariant int

const (
	// ButtonPrimary is the marketing-surface primary CTA. On Apple this is
	// the Action Blue pill; on Meta this is the black pill.
	ButtonPrimary ButtonVariant = iota
	// ButtonSecondary is the outlined ghost-pill alternative.
	ButtonSecondary
	// ButtonGhost is a quieter outlined or pearl-fill button for tertiary
	// actions.
	ButtonGhost
	// ButtonAccent is the commerce-flow CTA. On Meta this is cobalt; on
	// Apple it falls back to the Action Blue primary.
	ButtonAccent
	// ButtonOnDark is the primary CTA placed on a dark surface.
	ButtonOnDark
)

// Button renders one of the active theme's button styles. Provide either
// Label (plain text) or Child (any widget); OnPressed wires the click.
type Button struct {
	Variant   ButtonVariant
	Label     string
	Child     gutter.Widget
	OnPressed func()
}

func (b Button) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := buttonStyleFor(t, b.Variant)
	css := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"border-radius":    style.Rounded,
		"padding":          style.PaddingY + " " + style.PaddingX,
		"cursor":           "pointer",
		"display":          "inline-flex",
		"align-items":      "center",
		"justify-content":  "center",
		"text-align":       "center",
		"user-select":      "none",
		"transition":       "transform 0.15s ease-out, background-color 0.15s ease-out",
	}
	if style.BorderColor != "" && style.BorderWidth != "" {
		css["border"] = style.BorderWidth + " solid " + style.BorderColor
	} else {
		css["border"] = "none"
	}
	applySpec(css, style.Typography)

	w := Styled{Tag: "button", Style: css}
	if b.Child != nil {
		w.Children = []gutter.Widget{b.Child}
	} else {
		w.Text = b.Label
	}
	if b.OnPressed != nil {
		op := b.OnPressed
		w.Events = map[string]func(gutter.Event){
			"click": func(gutter.Event) { op() },
		}
	}
	return w
}

func buttonStyleFor(t *themes.Theme, variant ButtonVariant) themes.ButtonStyle {
	switch variant {
	case ButtonSecondary:
		return t.Components.ButtonSecondary
	case ButtonGhost:
		return t.Components.ButtonGhost
	case ButtonAccent:
		return t.Components.ButtonAccent
	case ButtonOnDark:
		return t.Components.ButtonOnDark
	default:
		return t.Components.ButtonPrimary
	}
}

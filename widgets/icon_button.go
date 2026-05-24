package widgets

import (
	"github.com/Runway-Club/gutter"
)

// IconButton is a button whose primary content is an [Icon]. It reuses the
// theme's button style — Variant picks the same palette as [Button] — but
// switches to square padding and a smaller form factor.
//
//	widgets.IconButton{
//	    Icon:      "favorite",
//	    Variant:   widgets.ButtonGhost,
//	    OnPressed: func() { liked.Set(!liked.Value()) },
//	}
type IconButton struct {
	// Icon is the Material Symbol name (e.g. "home", "favorite").
	Icon string
	// IconStyle picks the symbol family. Defaults to IconOutlined.
	IconStyle IconStyle
	// Filled toggles the FILL axis on the icon.
	Filled bool
	// Size is the icon's CSS size. Defaults to "24px".
	Size string
	// Tooltip is exposed as the button's title attribute for accessibility.
	Tooltip string
	// Variant picks the button palette.
	Variant   ButtonVariant
	OnPressed func()
}

func (b IconButton) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := buttonStyleFor(t, b.Variant)
	size := fallback(b.Size, "24px")

	css := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"border-radius":    fallback(style.Rounded, "9999px"),
		"padding":          "8px",
		"cursor":           "pointer",
		"display":          "inline-flex",
		"align-items":      "center",
		"justify-content":  "center",
		"user-select":      "none",
		"line-height":      "1",
		"transition":       "transform 0.15s ease-out, background-color 0.15s ease-out",
	}
	if style.BorderColor != "" && style.BorderWidth != "" {
		css["border"] = style.BorderWidth + " solid " + style.BorderColor
	} else {
		css["border"] = "none"
	}

	attrs := map[string]string{}
	if b.Tooltip != "" {
		attrs["title"] = b.Tooltip
		attrs["aria-label"] = b.Tooltip
	}

	icon := Icon{
		Name:   b.Icon,
		Size:   size,
		Color:  style.Foreground,
		Style:  b.IconStyle,
		Filled: b.Filled,
	}

	w := Styled{
		Tag:      "button",
		Attrs:    attrs,
		Style:    css,
		Children: []gutter.Widget{icon},
	}
	if b.OnPressed != nil {
		op := b.OnPressed
		w.Events = map[string]func(gutter.Event){
			"click": func(gutter.Event) { op() },
		}
	}
	return w
}

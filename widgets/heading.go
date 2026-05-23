package widgets

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

// HeadingLevel maps to one of the theme's display/heading typography roles.
// H1 is the hero display; H6 is the smallest section heading.
type HeadingLevel int

const (
	H1 HeadingLevel = iota + 1 // Typography.HeroDisplay
	H2                         // Typography.DisplayLarge
	H3                         // Typography.DisplayMedium
	H4                         // Typography.HeadingLarge
	H5                         // Typography.HeadingMedium
	H6                         // Typography.HeadingSmall
)

// Heading renders text in one of the active theme's heading roles. Color
// defaults to the theme's ink color; pass an explicit Color for placement on
// dark surfaces (e.g. ctx.Theme.Colors.OnDark).
type Heading struct {
	Level HeadingLevel
	Text  string
	Color string
}

func (h Heading) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	spec := headingSpec(t, h.Level)
	return Text{Data: h.Text, Style: styleFromSpec(spec, fallback(h.Color, t.Colors.Ink))}
}

func headingSpec(t *themes.Theme, level HeadingLevel) themes.TextSpec {
	switch level {
	case H1:
		return t.Typography.HeroDisplay
	case H2:
		return t.Typography.DisplayLarge
	case H3:
		return t.Typography.DisplayMedium
	case H4:
		return t.Typography.HeadingLarge
	case H5:
		return t.Typography.HeadingMedium
	case H6:
		return t.Typography.HeadingSmall
	default:
		return t.Typography.HeadingLarge
	}
}

// Body renders text in one of the theme's body roles. Bold flips to the
// strong variant; Small drops to caption size; both together gives the
// strong-caption role. Color defaults to the theme's ink color.
type Body struct {
	Text  string
	Bold  bool
	Small bool
	Color string
}

func (b Body) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	var spec themes.TextSpec
	switch {
	case b.Small && b.Bold:
		spec = t.Typography.CaptionStrong
	case b.Small:
		spec = t.Typography.Caption
	case b.Bold:
		spec = t.Typography.BodyStrong
	default:
		spec = t.Typography.Body
	}
	return Text{Data: b.Text, Style: styleFromSpec(spec, fallback(b.Color, t.Colors.Ink))}
}

// Caption is shorthand for Body{Small: true}.
type Caption struct {
	Text  string
	Bold  bool
	Color string
}

func (c Caption) Build(ctx *gutter.BuildContext) gutter.Widget {
	return Body{Text: c.Text, Bold: c.Bold, Small: true, Color: c.Color}.Build(ctx)
}

// Link renders an inline anchor styled with the theme's link typography.
// OnPressed wires a click handler; if nil, the link is non-interactive
// (still styled, e.g. inside a breadcrumb).
type Link struct {
	Text      string
	OnPressed func()
	Color     string
}

func (l Link) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := map[string]string{
		"color":           fallback(l.Color, t.Colors.Primary),
		"text-decoration": "none",
		"cursor":          "pointer",
	}
	applySpec(style, t.Typography.Link)
	w := Styled{
		Tag:   "a",
		Text:  l.Text,
		Style: style,
		Attrs: map[string]string{"href": "javascript:void(0)"},
	}
	if l.OnPressed != nil {
		op := l.OnPressed
		w.Events = map[string]func(gutter.Event){
			"click": func(gutter.Event) { op() },
		}
	}
	return w
}

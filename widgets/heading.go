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
	// Render a real <h1>–<h6> so screen readers and SEO see heading structure.
	// margin:0 drops the browser's default heading margins; the theme spec owns
	// sizing/weight.
	style := map[string]string{
		"color":  fallback(resolveColor(t, h.Color), t.Colors.Ink),
		"margin": "0",
	}
	applySpec(style, spec)
	return Styled{Tag: headingTag(h.Level), Text: h.Text, Style: style}
}

// headingTag maps a HeadingLevel to its semantic HTML tag.
func headingTag(level HeadingLevel) string {
	switch level {
	case H1:
		return "h1"
	case H2:
		return "h2"
	case H3:
		return "h3"
	case H4:
		return "h4"
	case H5:
		return "h5"
	case H6:
		return "h6"
	default:
		return "h2"
	}
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
	Text   string
	Bold   bool
	Small  bool
	Color  string
	Inline bool // render a <span> instead of a <p> so the text flows inline
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
	// Render a real <p> (margin reset to 0) so prose has paragraph semantics for
	// screen readers; the theme spec owns sizing/weight/color. Use Inline:true
	// for a <span> when the text must flow within a line.
	tag := "p"
	if b.Inline {
		tag = "span"
	}
	style := map[string]string{
		"color":  fallback(resolveColor(t, b.Color), t.Colors.Ink),
		"margin": "0",
	}
	applySpec(style, spec)
	return Styled{Tag: tag, Text: b.Text, Style: style}
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
	Text string
	// Href is a real URL for navigation. When set the anchor is a genuine,
	// crawlable link (good for SEO and accessibility); when empty the link is
	// JS-driven (OnPressed) and uses a no-op href.
	Href      string
	OnPressed func()
	Color     string
}

func (l Link) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := map[string]string{
		"color":           fallback(resolveColor(t, l.Color), t.Colors.Primary),
		"text-decoration": "none",
		"cursor":          "pointer",
	}
	applySpec(style, t.Typography.Link)
	href := l.Href
	if href == "" {
		href = "javascript:void(0)"
	}
	w := Styled{
		Tag:   "a",
		Text:  l.Text,
		Style: style,
		Attrs: map[string]string{"href": href},
	}
	if l.OnPressed != nil {
		op := l.OnPressed
		w.Events = map[string]func(gutter.Event){
			"click": func(gutter.Event) { op() },
		}
	}
	return w
}

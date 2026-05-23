package widgets

import (
	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

// SurfaceVariant selects which surface in theme.Components.Surface* to use.
type SurfaceVariant int

const (
	// SurfaceCanvas is the page-level canvas (Apple white, Meta white).
	SurfaceCanvas SurfaceVariant = iota
	// SurfaceAlt is the alternate light surface (Apple parchment, Meta soft
	// cloud). Used to break light tiles.
	SurfaceAlt
	// SurfaceDark is the dark tile / banner.
	SurfaceDark
)

// Surface is a themed layout region used for full-bleed sections, hero bands,
// and tile alternation. Padding defaults to the theme's section padding; pass
// Padding to override (use "0" for an edge-to-edge tile that controls its
// own child padding).
type Surface struct {
	Variant SurfaceVariant
	Padding string
	Child   gutter.Widget
}

func (s Surface) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := surfaceStyleFor(t, s.Variant)
	css := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"padding":          fallback(s.Padding, style.Padding),
		"width":            "100%",
		// height: 100% gives Surface a definite height — required so that
		// children using `height: 100%` (e.g. Center) can resolve their own
		// percentage against it.
		// min-height: 100% on top of that lets Surface grow past the
		// viewport when its content is larger, so the background still
		// covers the whole scrollable area.
		// For a Surface nested in an auto-height parent (e.g. a Column of
		// Surfaces, like examples/showcase), both percentages collapse and
		// Surface just sizes to its content — so nested Surfaces don't each
		// force a full viewport.
		"height":     "100%",
		"min-height": "100%",
		"box-sizing": "border-box",
	}
	if style.Rounded != "" {
		css["border-radius"] = style.Rounded
	}
	w := Styled{Style: css}
	if s.Child != nil {
		w.Children = []gutter.Widget{s.Child}
	}
	return w
}

func surfaceStyleFor(t *themes.Theme, variant SurfaceVariant) themes.SurfaceStyle {
	switch variant {
	case SurfaceAlt:
		return t.Components.SurfaceAlt
	case SurfaceDark:
		return t.Components.SurfaceDark
	default:
		return t.Components.SurfaceCanvas
	}
}

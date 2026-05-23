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

// styleFromSpec produces a TextStyle from a TextSpec plus an explicit color.
// Used by the typography widgets (Heading, Body, Caption).
func styleFromSpec(spec themes.TextSpec, color string) *TextStyle {
	return &TextStyle{
		Color:         color,
		FontFamily:    spec.FontFamily,
		FontSize:      spec.FontSize,
		FontWeight:    spec.FontWeight,
		LineHeight:    spec.LineHeight,
		LetterSpacing: spec.LetterSpacing,
	}
}

func fallback(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

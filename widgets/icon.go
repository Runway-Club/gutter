package widgets

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Runway-Club/gutter"
)

// IconStyle picks which Material Symbols family the Icon renders with.
// All three families are part of the same Google Fonts collection but each
// has its own class name and font file.
type IconStyle int

const (
	// IconOutlined uses the "Material Symbols Outlined" face (default).
	IconOutlined IconStyle = iota
	// IconRounded uses "Material Symbols Rounded".
	IconRounded
	// IconSharp uses "Material Symbols Sharp".
	IconSharp
)

// Icon renders a Google Material Symbols glyph. The host page must include
// the corresponding stylesheet (the gutter project template does this) — for
// the default Outlined style:
//
//	<link rel="stylesheet"
//	      href="https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined" />
//
// Pass Name as the symbol's identifier (lower-snake-case), e.g. "home",
// "favorite", "arrow_forward". The full glyph catalog is at
// https://fonts.google.com/icons. Color defaults to currentColor so the icon
// inherits the surrounding text color; Size defaults to "24px".
type Icon struct {
	Name   string
	Size   string
	Color  string
	Style  IconStyle
	Filled bool
	// Weight is the stroke weight (100..700). Zero leaves the family default
	// (400 for Outlined/Rounded, varies for Sharp).
	Weight int
	// Grade biases the stroke for visual emphasis (-25..200). Zero uses the
	// family default.
	Grade int
}

func (i Icon) Host() *gutter.Host {
	className := iconClass(i.Style)
	size := fallback(i.Size, "24px")
	style := map[string]string{
		"font-size":      size,
		"line-height":    "1",
		"display":        "inline-block",
		"vertical-align": "middle",
		"user-select":    "none",
	}
	if i.Color != "" {
		style["color"] = i.Color
	}
	style["font-variation-settings"] = iconVariation(i.Filled, i.Weight, i.Grade, size)
	return &gutter.Host{
		Tag:   "span",
		Text:  i.Name,
		Attrs: map[string]string{"class": className},
		Style: style,
	}
}

func iconClass(style IconStyle) string {
	switch style {
	case IconRounded:
		return "material-symbols-rounded"
	case IconSharp:
		return "material-symbols-sharp"
	default:
		return "material-symbols-outlined"
	}
}

// iconVariation builds the font-variation-settings string from the four
// Material Symbols axes (FILL, wght, GRAD, opsz). opsz mirrors the rendered
// font-size so the optical-size axis tracks the actual visual size.
func iconVariation(filled bool, weight, grade int, size string) string {
	fill := 0
	if filled {
		fill = 1
	}
	if weight == 0 {
		weight = 400
	}
	return fmt.Sprintf("'FILL' %d, 'wght' %d, 'GRAD' %d, 'opsz' %s",
		fill, weight, grade, parseSizePx(size))
}

// parseSizePx returns the numeric pixel value of size for the opsz axis.
// Best-effort: anything not parseable falls back to "24". Avoids fmt.Sscanf
// with %f — TinyGo's scanner panics on float conversion rather than erroring.
func parseSizePx(size string) string {
	s := strings.TrimSuffix(strings.TrimSpace(size), "px")
	if v, err := strconv.ParseFloat(strings.TrimSpace(s), 64); err == nil && v > 0 {
		return strconv.FormatFloat(v, 'g', -1, 64)
	}
	return "24"
}

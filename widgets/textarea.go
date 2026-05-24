package widgets

import (
	"fmt"

	"github.com/Runway-Club/gutter"
)

// TextArea is a multi-line text field. It reuses the theme's Input style for
// background/border/typography but ignores the fixed Height — the height
// comes from Rows (defaults to 4) and grows or scrolls based on Resize.
type TextArea struct {
	Value       string
	Placeholder string
	Rows        int
	Error       bool
	Disabled    bool
	ReadOnly    bool
	OnChanged   func(string)

	// Resize maps to the CSS resize property. Defaults to "vertical".
	// Accepts "none", "both", "horizontal", "vertical".
	Resize string

	// MaxLength caps the number of characters the user can type.
	MaxLength int

	Name string
}

func (a TextArea) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := t.Components.Input
	borderColor := style.BorderColor
	if a.Error {
		borderColor = style.BorderColorError
	}
	resize := a.Resize
	if resize == "" {
		resize = "vertical"
	}
	// InputStyle.Rounded is typically the theme's pill radius (e.g.
	// "9999px" on Apple), which is meant for single-line pill-shaped
	// inputs. A multi-line textarea with a 9999px radius is unreadable —
	// fall back to the theme's Large rounding (a sensible card-style
	// radius) for textareas.
	radius := t.Rounded.Large
	if radius == "" {
		radius = "12px"
	}
	css := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"border":           "1px solid " + borderColor,
		"border-radius":    radius,
		"padding":          style.Padding,
		"box-sizing":       "border-box",
		"outline":          "none",
		"resize":           resize,
		"width":            "100%",
		"min-height":       style.Height,
		"font-family":      "inherit",
	}
	if a.Disabled {
		css["opacity"] = "0.6"
		css["cursor"] = "not-allowed"
	}
	applySpec(css, style.Typography)

	rows := a.Rows
	if rows == 0 {
		rows = 4
	}
	attrs := map[string]string{
		"rows": fmt.Sprintf("%d", rows),
	}
	if a.Placeholder != "" {
		attrs["placeholder"] = a.Placeholder
	}
	if a.Disabled {
		attrs["disabled"] = ""
	}
	if a.ReadOnly {
		attrs["readonly"] = ""
	}
	if a.MaxLength > 0 {
		attrs["maxlength"] = fmt.Sprintf("%d", a.MaxLength)
	}
	if a.Name != "" {
		attrs["name"] = a.Name
	}

	value := a.Value
	w := propSyncHost{
		tag:   "textarea",
		attrs: attrs,
		style: css,
		onMount: func(node any) {
			// Same controlled-input dance as Input — set the property
			// directly (textarea's current value lives on `value`, not
			// textContent, after the user has interacted), and skip the
			// write when it would be a no-op to preserve caret position.
			setStringPropIfDifferent(node, "value", value)
		},
	}
	if a.OnChanged != nil {
		oc := a.OnChanged
		w.events = map[string]func(gutter.Event){
			"input": func(e gutter.Event) { oc(e.Value) },
		}
	}
	return w
}

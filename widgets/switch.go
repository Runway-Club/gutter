package widgets

import (
	"github.com/Runway-Club/gutter"
)

// Switch is a two-state toggle styled as a sliding pill. It's a button with
// role="switch" so screen readers announce the on/off state. Track and thumb
// transition with a 0.2s ease.
//
// Like Checkbox, Switch is controlled — the Checked field is the source of
// truth, OnChanged is invoked on toggle, and the parent is expected to
// rebuild with the new value.
type Switch struct {
	Checked   bool
	Label     string
	Disabled  bool
	OnChanged func(bool)
}

func (s Switch) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	trackOff := fallback(t.Colors.Hairline, "#d2d2d7")
	trackOn := fallback(t.Colors.Primary, "#0066cc")
	thumbColor := "#ffffff"

	track := trackOff
	thumbX := "2px"
	if s.Checked {
		track = trackOn
		thumbX = "22px"
	}

	trackStyle := map[string]string{
		"position":      "relative",
		"width":         "44px",
		"height":        "24px",
		"background":    track,
		"border-radius": "9999px",
		"border":        "none",
		"padding":       "0",
		"cursor":        "pointer",
		"transition":    "background-color 0.2s ease-out",
		"flex":          "none",
		"display":       "inline-block",
	}
	if s.Disabled {
		trackStyle["opacity"] = "0.6"
		trackStyle["cursor"] = "not-allowed"
	}
	thumbStyle := map[string]string{
		"position":      "absolute",
		"top":           "2px",
		"left":          thumbX,
		"width":         "20px",
		"height":        "20px",
		"background":    thumbColor,
		"border-radius": "50%",
		"transition":    "left 0.2s ease-out",
		"box-shadow":    "0 1px 3px rgba(0,0,0,0.2)",
	}

	attrs := map[string]string{
		"type":         "button",
		"role":         "switch",
		"aria-checked": boolAttr(s.Checked),
	}
	if s.Disabled {
		attrs["disabled"] = ""
	}

	checked := s.Checked
	events := map[string]func(gutter.Event){}
	if s.OnChanged != nil && !s.Disabled {
		oc := s.OnChanged
		events["click"] = func(gutter.Event) { oc(!checked) }
	}

	toggle := Styled{
		Tag:      "button",
		Attrs:    attrs,
		Style:    trackStyle,
		Events:   events,
		Children: []gutter.Widget{Styled{Style: thumbStyle}},
	}

	if s.Label == "" {
		return toggle
	}

	labelStyle := map[string]string{
		"display":     "inline-flex",
		"align-items": "center",
		"gap":         "12px",
		"user-select": "none",
		"color":       t.Colors.Ink,
	}
	if s.Disabled {
		labelStyle["opacity"] = "0.6"
	}
	applySpec(labelStyle, t.Components.Input.Typography)

	return Styled{
		Tag:      "label",
		Style:    labelStyle,
		Children: []gutter.Widget{toggle, Styled{Text: s.Label}},
	}
}

func boolAttr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

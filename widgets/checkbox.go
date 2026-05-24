package widgets

import (
	"github.com/Runway-Club/gutter"
)

// Checkbox is a single boolean toggle rendered as a native <input
// type="checkbox"> wrapped in a label. The browser's checkbox is themed
// through the CSS `accent-color` property, which picks up the active
// theme's primary color.
//
//	checked := gutter.NewNotifier(false)
//	widgets.ObserverBuilder[bool]{
//	    Source: checked,
//	    Builder: func(_ *gutter.BuildContext, v bool) gutter.Widget {
//	        return widgets.Checkbox{
//	            Checked:   v,
//	            Label:     "I agree",
//	            OnChanged: func(b bool) { checked.Set(b) },
//	        }
//	    },
//	}
type Checkbox struct {
	Checked   bool
	Label     string
	Disabled  bool
	OnChanged func(bool)
	Name      string
}

func (c Checkbox) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	labelCSS := map[string]string{
		"display":      "inline-flex",
		"align-items":  "center",
		"gap":          "8px",
		"cursor":       "pointer",
		"user-select":  "none",
		"color":        t.Colors.Ink,
		"accent-color": t.Colors.Primary,
	}
	if c.Disabled {
		labelCSS["opacity"] = "0.6"
		labelCSS["cursor"] = "not-allowed"
	}
	applySpec(labelCSS, t.Components.Input.Typography)

	attrs := map[string]string{"type": "checkbox"}
	if c.Disabled {
		attrs["disabled"] = ""
	}
	if c.Name != "" {
		attrs["name"] = c.Name
	}
	if c.Checked {
		attrs["checked"] = ""
	}

	checked := c.Checked
	input := propSyncHost{
		tag:   "input",
		attrs: attrs,
		style: map[string]string{
			"width":  "18px",
			"height": "18px",
			"margin": "0",
		},
		onMount: func(node any) { setBoolProp(node, "checked", checked) },
	}
	if c.OnChanged != nil {
		oc := c.OnChanged
		input.events = map[string]func(gutter.Event){
			"change": func(e gutter.Event) {
				// For checkboxes, e.Value is "on"/"" — not useful. Read the
				// checked property via the next rebuild's prop sync; for
				// now compute from current state.
				oc(!checked)
			},
		}
	}

	children := []gutter.Widget{input}
	if c.Label != "" {
		children = append(children, Styled{Text: c.Label})
	}

	return Styled{
		Tag:      "label",
		Style:    labelCSS,
		Children: children,
	}
}

package widgets

import "github.com/Runway-Club/gutter"

// Input renders a themed text field. Set Error to switch to the error-border
// variant. OnChanged fires for every keystroke (the DOM "input" event).
//
// The element-tree reconciler updates the input in place, so focus is
// preserved across rebuilds.
type Input struct {
	Value       string
	Placeholder string
	Error       bool
	OnChanged   func(string)
}

func (i Input) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := t.Components.Input
	borderColor := style.BorderColor
	borderWidth := "1px"
	if i.Error {
		borderColor = style.BorderColorError
	}
	css := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"border":           borderWidth + " solid " + borderColor,
		"border-radius":    style.Rounded,
		"padding":          style.Padding,
		"height":           style.Height,
		"box-sizing":       "border-box",
		"outline":          "none",
	}
	applySpec(css, style.Typography)
	attrs := map[string]string{"type": "text"}
	if i.Value != "" {
		attrs["value"] = i.Value
	}
	if i.Placeholder != "" {
		attrs["placeholder"] = i.Placeholder
	}
	w := Styled{Tag: "input", Attrs: attrs, Style: css}
	if i.OnChanged != nil {
		oc := i.OnChanged
		w.Events = map[string]func(gutter.Event){
			"input": func(e gutter.Event) { oc(e.Value) },
		}
	}
	return w
}

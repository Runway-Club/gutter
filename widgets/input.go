package widgets

import "github.com/Runway-Club/gutter"

// InputType selects the HTML <input type=...> attribute. The styling stays
// the same across all variants (it comes from theme.Components.Input); only
// the keyboard layout, validation, and platform widget shown by the browser
// change.
type InputType string

const (
	InputText          InputType = "text"
	InputPassword      InputType = "password"
	InputEmail         InputType = "email"
	InputNumber        InputType = "number"
	InputTel           InputType = "tel"
	InputURL           InputType = "url"
	InputSearch        InputType = "search"
	InputDate          InputType = "date"
	InputTime          InputType = "time"
	InputDateTimeLocal InputType = "datetime-local"
	InputMonth         InputType = "month"
	InputWeek          InputType = "week"
	InputColor         InputType = "color"
)

// Input renders a themed text field. Type picks the HTML input variant
// (defaults to InputText). Set Error to switch to the error-border style.
// OnChanged fires for every keystroke (the DOM "input" event).
//
// The element-tree reconciler updates the input in place, so focus is
// preserved across rebuilds. Value is driven through the DOM `value`
// property (via OnMount) rather than the `value` attribute: setAttribute
// on a focused input causes focus churn that scrolls the page to the
// focused element on every keystroke. Setting the property is silent and
// only mutates the visible content when the new value actually differs.
type Input struct {
	Type        InputType
	Value       string
	Placeholder string
	Error       bool
	Disabled    bool
	ReadOnly    bool
	OnChanged   func(string)

	// Numeric/date constraints. Empty values are omitted from the rendered
	// element. Step accepts "any" too, matching the HTML spec.
	Min, Max, Step string

	// Pattern is a regex for client-side validation (text-like types only).
	Pattern string
	// AutoComplete maps directly to the HTML attribute (e.g. "off", "email",
	// "current-password").
	AutoComplete string
	// Name is the form field name for native form submission.
	Name string
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
		"accent-color":     t.Colors.Primary,
	}
	if i.Disabled {
		css["opacity"] = "0.6"
		css["cursor"] = "not-allowed"
	}
	applySpec(css, style.Typography)

	typ := i.Type
	if typ == "" {
		typ = InputText
	}
	attrs := map[string]string{"type": string(typ)}
	if i.Placeholder != "" {
		attrs["placeholder"] = i.Placeholder
	}
	if i.Disabled {
		attrs["disabled"] = ""
	}
	if i.ReadOnly {
		attrs["readonly"] = ""
	}
	if i.Min != "" {
		attrs["min"] = i.Min
	}
	if i.Max != "" {
		attrs["max"] = i.Max
	}
	if i.Step != "" {
		attrs["step"] = i.Step
	}
	if i.Pattern != "" {
		attrs["pattern"] = i.Pattern
	}
	if i.AutoComplete != "" {
		attrs["autocomplete"] = i.AutoComplete
	}
	if i.Name != "" {
		attrs["name"] = i.Name
	}

	value := i.Value
	w := propSyncHost{
		tag:   "input",
		attrs: attrs,
		style: css,
		onMount: func(node any) {
			// Set the property only when it actually differs from what
			// the DOM holds. Writing on every reconcile would move the
			// caret to the end of the value on every keystroke, which is
			// the user-visible symptom of a "controlled input" done
			// naively. Skipping the no-op write preserves caret position
			// for the common case (user typed → state echoed back).
			setStringPropIfDifferent(node, "value", value)
		},
	}
	if i.OnChanged != nil {
		oc := i.OnChanged
		w.events = map[string]func(gutter.Event){
			"input": func(e gutter.Event) { oc(e.Value) },
		}
	}
	return w
}

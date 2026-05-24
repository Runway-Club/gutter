package widgets

import (
	"strconv"

	"github.com/Runway-Club/gutter"
)

// SelectOption is one entry in a Select dropdown. Label is what the user
// sees; Value is the typed value returned through OnChanged.
type SelectOption[T any] struct {
	Value T
	Label string
}

// Select is a themed <select> dropdown over a strongly-typed option set.
// The HTML option value is the index in Options, so any comparable T works
// — strings, ints, custom structs — without needing a Stringer.
//
//	type Color int
//	const ( Red Color = iota; Green; Blue )
//
//	widgets.Select[Color]{
//	    Options: []widgets.SelectOption[Color]{
//	        {Value: Red,   Label: "Red"},
//	        {Value: Green, Label: "Green"},
//	        {Value: Blue,  Label: "Blue"},
//	    },
//	    Selected:  current,
//	    OnChanged: func(c Color) { current = c },
//	}
//
// If Placeholder is set, an initial disabled option with that text and an
// empty value is prepended; it disappears once the user picks anything.
type Select[T comparable] struct {
	Options     []SelectOption[T]
	Selected    T
	OnChanged   func(T)
	Disabled    bool
	Placeholder string
	Name        string
}

func (s Select[T]) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := t.Components.Input
	css := map[string]string{
		"background-color": style.Background,
		"color":            style.Foreground,
		"border":           "1px solid " + style.BorderColor,
		"border-radius":    style.Rounded,
		"padding":          style.Padding,
		"height":           style.Height,
		"box-sizing":       "border-box",
		"outline":          "none",
		"accent-color":     t.Colors.Primary,
		"cursor":           "pointer",
		"appearance":       "auto",
		// The visible value of a <select> sits to the left of the native
		// chevron, which makes the text look uncentered next to a
		// symmetrically-padded box. text-align-last centers the
		// collapsed display without touching the dropdown options (which
		// stay left-aligned for readability).
		"text-align-last": "center",
	}
	// Override the Typography's line-height for selects. The theme's
	// input line-height (1.47 × 17px ≈ 25px) is taller than the content
	// box (44px height minus 24px vertical padding = 20px), which makes
	// the visible value sit slightly above center in some browsers.
	// Forcing line-height to 1 lets the browser vertically center the
	// glyph inside the padded box.
	applySpec(css, style.Typography)
	css["line-height"] = "1"
	if s.Disabled {
		css["opacity"] = "0.6"
		css["cursor"] = "not-allowed"
	}

	selectedIndex := -1
	for i, opt := range s.Options {
		if opt.Value == s.Selected {
			selectedIndex = i
			break
		}
	}

	children := []gutter.Widget{}
	if s.Placeholder != "" {
		placeholderAttrs := map[string]string{"value": "", "disabled": ""}
		if selectedIndex < 0 {
			placeholderAttrs["selected"] = ""
		}
		children = append(children, Styled{
			Tag:   "option",
			Attrs: placeholderAttrs,
			Text:  s.Placeholder,
		})
	}
	for i, opt := range s.Options {
		attrs := map[string]string{"value": strconv.Itoa(i)}
		if i == selectedIndex {
			attrs["selected"] = ""
		}
		children = append(children, Styled{
			Tag:   "option",
			Attrs: attrs,
			Text:  opt.Label,
		})
	}

	attrs := map[string]string{}
	if s.Disabled {
		attrs["disabled"] = ""
	}
	if s.Name != "" {
		attrs["name"] = s.Name
	}

	idxValue := ""
	if selectedIndex >= 0 {
		idxValue = strconv.Itoa(selectedIndex)
	}

	dispatch := s.OnChanged
	options := s.Options
	host := propSyncHost{
		tag:      "select",
		attrs:    attrs,
		style:    css,
		children: children,
		onMount: func(node any) {
			// `selected` attribute is only honored at parse time; once the
			// user picks, setAttribute("selected") on a different option
			// doesn't update the rendered selection. Setting the DOM
			// `value` property does.
			setStringProp(node, "value", idxValue)
		},
	}
	if dispatch != nil {
		host.events = map[string]func(gutter.Event){
			"change": func(e gutter.Event) {
				i, err := strconv.Atoi(e.Value)
				if err != nil || i < 0 || i >= len(options) {
					return
				}
				dispatch(options[i].Value)
			},
		}
	}
	return host
}

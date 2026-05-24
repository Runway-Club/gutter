package widgets

import (
	"fmt"
	"strconv"

	"github.com/Runway-Club/gutter"
)

// Slider is a horizontal range input rendered as a native <input
// type="range">. The browser draws the track and thumb; `accent-color`
// tints them with the active theme's primary color.
//
// Min defaults to 0, Max defaults to 100, Step to 1. Value is clamped by
// the browser to the [Min, Max] range. OnChanged fires for every drag
// position (the DOM "input" event).
type Slider struct {
	Value     float64
	Min       float64
	Max       float64
	Step      float64
	Disabled  bool
	OnChanged func(float64)
	Name      string
}

func (s Slider) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	min := s.Min
	max := s.Max
	if min == 0 && max == 0 {
		max = 100
	}
	step := s.Step
	if step == 0 {
		step = 1
	}
	css := map[string]string{
		"width":        "100%",
		"accent-color": t.Colors.Primary,
		"cursor":       "pointer",
	}
	if s.Disabled {
		css["opacity"] = "0.6"
		css["cursor"] = "not-allowed"
	}

	attrs := map[string]string{
		"type":  "range",
		"min":   fmt.Sprintf("%g", min),
		"max":   fmt.Sprintf("%g", max),
		"step":  fmt.Sprintf("%g", step),
		"value": fmt.Sprintf("%g", s.Value),
	}
	if s.Disabled {
		attrs["disabled"] = ""
	}
	if s.Name != "" {
		attrs["name"] = s.Name
	}

	value := s.Value
	input := propSyncHost{
		tag:   "input",
		attrs: attrs,
		style: css,
		onMount: func(node any) {
			setStringProp(node, "value", fmt.Sprintf("%g", value))
		},
	}
	if s.OnChanged != nil {
		oc := s.OnChanged
		input.events = map[string]func(gutter.Event){
			"input": func(e gutter.Event) {
				v, err := strconv.ParseFloat(e.Value, 64)
				if err != nil {
					return
				}
				oc(v)
			},
		}
	}
	return input
}

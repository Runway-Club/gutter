package widgets

import (
	"fmt"

	"github.com/Runway-Club/gutter"
)

// MainAxis* values map directly to CSS justify-content keywords. Use the
// constants for clarity; raw CSS values are also accepted.
const (
	MainAxisStart        = "flex-start"
	MainAxisCenter       = "center"
	MainAxisEnd          = "flex-end"
	MainAxisSpaceBetween = "space-between"
	MainAxisSpaceAround  = "space-around"
	MainAxisSpaceEvenly  = "space-evenly"
)

// CrossAxis* values map to CSS align-items keywords.
const (
	CrossAxisStart    = "flex-start"
	CrossAxisCenter   = "center"
	CrossAxisEnd      = "flex-end"
	CrossAxisStretch  = "stretch"
	CrossAxisBaseline = "baseline"
)

// Column lays its children out vertically using flexbox.
type Column struct {
	Children       []gutter.Widget
	MainAxisAlign  string
	CrossAxisAlign string
	Spacing        float64
}

func (c Column) Host() *gutter.Host {
	return flexHost("column", c.Children, c.MainAxisAlign, c.CrossAxisAlign, c.Spacing)
}

// Row lays its children out horizontally using flexbox.
type Row struct {
	Children       []gutter.Widget
	MainAxisAlign  string
	CrossAxisAlign string
	Spacing        float64
}

func (r Row) Host() *gutter.Host {
	return flexHost("row", r.Children, r.MainAxisAlign, r.CrossAxisAlign, r.Spacing)
}

func flexHost(direction string, children []gutter.Widget, mainAxis, crossAxis string, spacing float64) *gutter.Host {
	h := &gutter.Host{Tag: "div", Style: map[string]string{
		"display":        "flex",
		"flex-direction": direction,
	}}
	if mainAxis != "" {
		h.Style["justify-content"] = mainAxis
	}
	if crossAxis != "" {
		h.Style["align-items"] = crossAxis
	}
	if spacing > 0 {
		h.Style["gap"] = fmt.Sprintf("%gpx", spacing)
	}
	h.Children = children
	return h
}

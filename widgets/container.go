package widgets

import (
	"fmt"

	"github.com/Runway-Club/gutter"
)

// EdgeInsets is a uniform spacing descriptor used for padding and margin.
// Values are in pixels.
type EdgeInsets struct {
	Top, Right, Bottom, Left float64
}

// EdgeInsetsAll returns the same value for all four sides.
func EdgeInsetsAll(v float64) EdgeInsets {
	return EdgeInsets{v, v, v, v}
}

// EdgeInsetsSymmetric returns insets with matching vertical and horizontal values.
func EdgeInsetsSymmetric(vertical, horizontal float64) EdgeInsets {
	return EdgeInsets{vertical, horizontal, vertical, horizontal}
}

// IsZero reports whether all four sides are zero.
func (e EdgeInsets) IsZero() bool {
	return e.Top == 0 && e.Right == 0 && e.Bottom == 0 && e.Left == 0
}

// CSS renders the insets as a CSS top/right/bottom/left shorthand string.
func (e EdgeInsets) CSS() string {
	return fmt.Sprintf("%gpx %gpx %gpx %gpx", e.Top, e.Right, e.Bottom, e.Left)
}

// Container is a styled <div> with a single optional child. Use it as a
// general-purpose layout box for background color, padding, sizing, etc.
type Container struct {
	Child        gutter.Widget
	Padding      EdgeInsets
	Margin       EdgeInsets
	Color        string
	Width        string
	Height       string
	BorderRadius string
	Border       string
}

func (c Container) Host() *gutter.Host {
	h := &gutter.Host{Tag: "div", Style: map[string]string{}}
	if !c.Padding.IsZero() {
		h.Style["padding"] = c.Padding.CSS()
	}
	if !c.Margin.IsZero() {
		h.Style["margin"] = c.Margin.CSS()
	}
	if c.Color != "" {
		h.Style["background-color"] = c.Color
	}
	if c.Width != "" {
		h.Style["width"] = c.Width
	}
	if c.Height != "" {
		h.Style["height"] = c.Height
	}
	if c.BorderRadius != "" {
		h.Style["border-radius"] = c.BorderRadius
	}
	if c.Border != "" {
		h.Style["border"] = c.Border
	}
	if c.Child != nil {
		h.Children = []gutter.Widget{c.Child}
	}
	return h
}

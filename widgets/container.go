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

// Container is the general-purpose layout box: a styled <div> with one
// optional child. It covers the styling most screens need — background,
// spacing, sizing, borders, shadow, positioning, overflow — so app code rarely
// has to drop down to Styled and raw CSS.
//
// Color and BorderColor accept either a raw CSS color or a theme Color token
// (ColorPrimary, ColorInk, …), resolved against the active theme at build
// time. Container is a StatelessWidget for exactly this reason: a HostWidget
// can't see ctx.Theme, so it could never reference theme colors by role.
//
// Most fields map to the obvious CSS property and are omitted when empty.
type Container struct {
	Child gutter.Widget

	// Spacing.
	Padding EdgeInsets
	Margin  EdgeInsets
	Gap     float64 // gap between children when AlignChildren lays them out

	// Background and border. Color/BorderColor take a raw CSS color or a
	// theme Color token. Border (a full CSS shorthand) wins over the
	// BorderColor/BorderWidth pair.
	Color        string
	BorderRadius string
	Border       string
	BorderColor  string
	BorderWidth  string // defaults to "1px" when BorderColor is set

	// Sizing.
	Width     string
	Height    string
	MinWidth  string
	MaxWidth  string
	MinHeight string
	MaxHeight string

	// Box behaviour.
	Shadow   string // box-shadow
	Overflow string // overflow
	Cursor   string // cursor
	Opacity  string // opacity ("0".."1"); raw string so 0 is expressible

	// Positioning. Position is "relative"/"absolute"/"fixed"/"sticky"; the
	// inset fields are applied when set.
	Position                 string
	Top, Right, Bottom, Left string
	ZIndex                   string

	// Flex participation, for when this Container is a child of Row/Column.
	Flex      string // flex shorthand, e.g. "1"
	AlignSelf string

	// Transition shorthand for animating the above on state changes.
	Transition string
}

func (c Container) Build(ctx *gutter.BuildContext) gutter.Widget {
	t := activeTheme(ctx)
	style := map[string]string{}

	if !c.Padding.IsZero() {
		style["padding"] = c.Padding.CSS()
	}
	if !c.Margin.IsZero() {
		style["margin"] = c.Margin.CSS()
	}
	if c.Gap > 0 {
		style["gap"] = px(c.Gap)
	}

	setIf(style, "background-color", resolveColor(t, c.Color))
	setIf(style, "border-radius", c.BorderRadius)
	switch {
	case c.Border != "":
		style["border"] = c.Border
	case c.BorderColor != "":
		width := fallback(c.BorderWidth, "1px")
		style["border"] = width + " solid " + resolveColor(t, c.BorderColor)
	}

	setIf(style, "width", c.Width)
	setIf(style, "height", c.Height)
	setIf(style, "min-width", c.MinWidth)
	setIf(style, "max-width", c.MaxWidth)
	setIf(style, "min-height", c.MinHeight)
	setIf(style, "max-height", c.MaxHeight)

	setIf(style, "box-shadow", c.Shadow)
	setIf(style, "overflow", c.Overflow)
	setIf(style, "cursor", c.Cursor)
	setIf(style, "opacity", c.Opacity)

	setIf(style, "position", c.Position)
	setIf(style, "top", c.Top)
	setIf(style, "right", c.Right)
	setIf(style, "bottom", c.Bottom)
	setIf(style, "left", c.Left)
	setIf(style, "z-index", c.ZIndex)

	setIf(style, "flex", c.Flex)
	setIf(style, "align-self", c.AlignSelf)
	setIf(style, "transition", c.Transition)

	var children []gutter.Widget
	if c.Child != nil {
		children = []gutter.Widget{c.Child}
	}
	return Styled{Style: style, Children: children}
}

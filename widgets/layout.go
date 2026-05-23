package widgets

import "github.com/Runway-Club/gutter"

// Padding wraps its child in a <div> with the given padding.
type Padding struct {
	Padding EdgeInsets
	Child   gutter.Widget
}

func (p Padding) Host() *gutter.Host {
	h := &gutter.Host{Tag: "div", Style: map[string]string{"padding": p.Padding.CSS()}}
	if p.Child != nil {
		h.Children = []gutter.Widget{p.Child}
	}
	return h
}

// Center horizontally and vertically centers its child inside a full-size box.
type Center struct {
	Child gutter.Widget
}

func (c Center) Host() *gutter.Host {
	h := &gutter.Host{Tag: "div", Style: map[string]string{
		"display":         "flex",
		"justify-content": "center",
		"align-items":     "center",
		"width":           "100%",
		"height":          "100%",
	}}
	if c.Child != nil {
		h.Children = []gutter.Widget{c.Child}
	}
	return h
}

// SizedBox forces a child into a fixed width/height. Empty values let CSS pick.
type SizedBox struct {
	Width  string
	Height string
	Child  gutter.Widget
}

func (s SizedBox) Host() *gutter.Host {
	h := &gutter.Host{Tag: "div", Style: map[string]string{}}
	if s.Width != "" {
		h.Style["width"] = s.Width
	}
	if s.Height != "" {
		h.Style["height"] = s.Height
	}
	if s.Child != nil {
		h.Children = []gutter.Widget{s.Child}
	}
	return h
}

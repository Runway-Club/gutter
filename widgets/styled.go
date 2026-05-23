package widgets

import "github.com/Runway-Club/gutter"

// Styled is the escape hatch for arbitrary CSS / DOM construction. It renders
// the given Tag (defaults to "div") with the supplied attrs, style, events,
// children, and optional text content. Themed widgets build on top of this;
// apps rarely need to use it directly.
type Styled struct {
	Tag      string
	Text     string
	Attrs    map[string]string
	Style    map[string]string
	Events   map[string]func(gutter.Event)
	Children []gutter.Widget
}

func (s Styled) Host() *gutter.Host {
	tag := s.Tag
	if tag == "" {
		tag = "div"
	}
	return &gutter.Host{
		Tag:      tag,
		Text:     s.Text,
		Attrs:    s.Attrs,
		Style:    s.Style,
		Events:   s.Events,
		Children: s.Children,
	}
}

package widgets

import "github.com/Runway-Club/gutter"

// Text renders a string. Style is optional; when non-nil its non-empty fields
// are translated to inline CSS.
type Text struct {
	Data  string
	Style *TextStyle
}

// TextStyle is the inline-CSS shape for typography. All fields are CSS values;
// empty fields are omitted from the rendered style. Themed widgets fill these
// from a TextSpec in the active theme so apps don't have to.
type TextStyle struct {
	Color         string
	FontFamily    string
	FontSize      string
	FontWeight    string
	LineHeight    string
	LetterSpacing string
}

func (t Text) Host() *gutter.Host {
	h := &gutter.Host{Tag: "span", Text: t.Data}
	if t.Style != nil {
		h.Style = map[string]string{}
		setIf(h.Style, "color", t.Style.Color)
		setIf(h.Style, "font-family", t.Style.FontFamily)
		setIf(h.Style, "font-size", t.Style.FontSize)
		setIf(h.Style, "font-weight", t.Style.FontWeight)
		setIf(h.Style, "line-height", t.Style.LineHeight)
		setIf(h.Style, "letter-spacing", t.Style.LetterSpacing)
	}
	return h
}

func setIf(m map[string]string, key, value string) {
	if value != "" {
		m[key] = value
	}
}

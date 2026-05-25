package widgets

import (
	"strings"
	"testing"

	"github.com/Runway-Club/gutter"
)

// Proves the real themed catalog renders to HTML on the host platform: Build()
// runs, theme tokens resolve, and no syscall/js is touched.
func TestSSRRendersThemedCatalog(t *testing.T) {
	app := Scaffold{
		Title: "Demo",
		Body: Column{Children: []gutter.Widget{
			Heading{Level: H1, Text: "Hello"},
			Body{Text: "some body text"},
			Button{Variant: ButtonPrimary, Label: "Click me"},
		}},
	}
	out, err := gutter.RenderToHTML(app)
	if err != nil {
		t.Fatalf("RenderToHTML: %v", err)
	}
	if !strings.HasPrefix(out, "<") {
		t.Fatalf("expected HTML, got: %q", out)
	}
	for _, want := range []string{"Hello", "some body text", "Click me", "style=", "background-color"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in SSR output:\n%s", want, out)
		}
	}
}

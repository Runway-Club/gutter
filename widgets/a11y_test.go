package widgets

import (
	"testing"

	"github.com/Runway-Club/gutter/themes"
)

// Body is prose, so it renders a <p> (paragraph semantics for screen readers)
// with the browser's default margin reset; Inline opts back into a <span>.
func TestBodyRendersParagraph(t *testing.T) {
	h := hostOf(t, Body{Text: "x"})
	wantTag(t, h, "p")
	wantStyle(t, h, "margin", "0")

	inline := hostOf(t, Body{Text: "x", Inline: true})
	wantTag(t, inline, "span")
}

// Scaffold wraps Body in a <main> landmark and Footer in a <footer> landmark
// (AppBar already renders <header>) so assistive tech can navigate by region.
func TestScaffoldLandmarks(t *testing.T) {
	root := hostOfCtx(t, Scaffold{
		Body:   Body{Text: "content"},
		Footer: Caption{Text: "foot"},
	}, testCtx(themes.Apple))

	var tags []string
	for _, c := range root.Children {
		tags = append(tags, hostOf(t, c).Tag)
	}
	if len(tags) != 2 || tags[0] != "main" || tags[1] != "footer" {
		t.Fatalf("landmark tags = %v, want [main footer]", tags)
	}
}

func TestDialogAttrs(t *testing.T) {
	open := dialogAttrs(true)
	if open["role"] != "dialog" || open["aria-modal"] != "true" {
		t.Fatalf("open dialog attrs = %v", open)
	}
	if _, ok := open["aria-hidden"]; ok {
		t.Errorf("an open dialog must not be aria-hidden")
	}
	if dialogAttrs(false)["aria-hidden"] != "true" {
		t.Errorf("a closed dialog must be aria-hidden")
	}
}

// The dialog attrs must actually reach the overlay sheet.
func TestPopupSheetHasDialogRole(t *testing.T) {
	ctx := testCtx(themes.Apple)
	openSheet := popupRender(ctx, Popup{Child: Text{Data: "x"}}, true).(Styled).Children[1].(Styled)
	if openSheet.Attrs["role"] != "dialog" {
		t.Fatalf("popup sheet role = %q, want dialog", openSheet.Attrs["role"])
	}
	closedSheet := popupRender(ctx, Popup{Child: Text{Data: "x"}}, false).(Styled).Children[1].(Styled)
	if closedSheet.Attrs["aria-hidden"] != "true" {
		t.Errorf("closed popup sheet should be aria-hidden")
	}
}

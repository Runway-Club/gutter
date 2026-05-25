package widgets

import (
	"testing"

	"github.com/Runway-Club/gutter"
)

func TestText(t *testing.T) {
	h := hostOf(t, Text{Data: "hello"})
	wantTag(t, h, "span")
	if h.Text != "hello" {
		t.Errorf("Text = %q, want %q", h.Text, "hello")
	}
	// No style map when Style is nil.
	if h.Style != nil {
		t.Errorf("nil TextStyle should leave Host.Style nil, got %v", h.Style)
	}
}

func TestTextStyleMapsToCSS(t *testing.T) {
	h := hostOf(t, Text{Data: "x", Style: &TextStyle{
		Color:      "#111",
		FontSize:   "14px",
		FontWeight: "600",
	}})
	wantStyle(t, h, "color", "#111")
	wantStyle(t, h, "font-size", "14px")
	wantStyle(t, h, "font-weight", "600")
	// Empty fields are omitted.
	wantNoStyle(t, h, "line-height")
	wantNoStyle(t, h, "letter-spacing")
}

func TestColumnFlex(t *testing.T) {
	h := hostOf(t, Column{
		MainAxisAlign:  MainAxisSpaceBetween,
		CrossAxisAlign: CrossAxisCenter,
		Spacing:        12,
		Children:       []gutter.Widget{Text{Data: "a"}, Text{Data: "b"}},
	})
	wantTag(t, h, "div")
	wantStyle(t, h, "display", "flex")
	wantStyle(t, h, "flex-direction", "column")
	wantStyle(t, h, "justify-content", "space-between")
	wantStyle(t, h, "align-items", "center")
	wantStyle(t, h, "gap", "12px")
	wantChildren(t, h, 2)
}

func TestRowFlexDirection(t *testing.T) {
	h := hostOf(t, Row{Children: []gutter.Widget{Text{Data: "a"}}})
	wantStyle(t, h, "display", "flex")
	wantStyle(t, h, "flex-direction", "row")
	// No alignment / spacing → those properties stay unset.
	wantNoStyle(t, h, "justify-content")
	wantNoStyle(t, h, "align-items")
	wantNoStyle(t, h, "gap")
}

func TestCenterFillsAndCenters(t *testing.T) {
	h := hostOf(t, Center{Child: Text{Data: "x"}})
	wantStyle(t, h, "display", "flex")
	wantStyle(t, h, "justify-content", "center")
	wantStyle(t, h, "align-items", "center")
	wantStyle(t, h, "width", "100%")
	wantStyle(t, h, "height", "100%")
	wantChildren(t, h, 1)
}

func TestCenterNilChild(t *testing.T) {
	h := hostOf(t, Center{})
	wantChildren(t, h, 0)
}

func TestPadding(t *testing.T) {
	h := hostOf(t, Padding{Padding: EdgeInsetsAll(8), Child: Text{Data: "x"}})
	wantStyle(t, h, "padding", "8px 8px 8px 8px")
	wantChildren(t, h, 1)
}

func TestSizedBox(t *testing.T) {
	h := hostOf(t, SizedBox{Width: "100px", Child: Text{Data: "x"}})
	wantStyle(t, h, "width", "100px")
	wantNoStyle(t, h, "height")
	wantChildren(t, h, 1)
}

func TestStyledDefaultsToDiv(t *testing.T) {
	h := hostOf(t, Styled{})
	wantTag(t, h, "div")
}

func TestStyledPassesThrough(t *testing.T) {
	h := hostOf(t, Styled{
		Tag:   "section",
		Text:  "body",
		Attrs: map[string]string{"data-testid": "x"},
		Style: map[string]string{"color": "red"},
	})
	wantTag(t, h, "section")
	if h.Text != "body" {
		t.Errorf("Text = %q, want %q", h.Text, "body")
	}
	if h.Attrs["data-testid"] != "x" {
		t.Errorf("attr data-testid = %q, want x", h.Attrs["data-testid"])
	}
	wantStyle(t, h, "color", "red")
}

func TestEdgeInsets(t *testing.T) {
	if got := EdgeInsetsAll(4).CSS(); got != "4px 4px 4px 4px" {
		t.Errorf("EdgeInsetsAll(4).CSS() = %q", got)
	}
	if got := EdgeInsetsSymmetric(8, 16).CSS(); got != "8px 16px 8px 16px" {
		t.Errorf("EdgeInsetsSymmetric(8,16).CSS() = %q", got)
	}
	if !(EdgeInsets{}).IsZero() {
		t.Error("zero EdgeInsets should be IsZero")
	}
	if (EdgeInsets{Top: 1}).IsZero() {
		t.Error("non-zero EdgeInsets should not be IsZero")
	}
}

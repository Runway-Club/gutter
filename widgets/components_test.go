package widgets

import (
	"strings"
	"testing"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

func TestInputDefaultType(t *testing.T) {
	h := hostOfCtx(t, Input{Placeholder: "Name"}, testCtx(themes.Apple))
	wantTag(t, h, "input")
	if h.Attrs["type"] != "text" {
		t.Errorf("default input type = %q, want text", h.Attrs["type"])
	}
	if h.Attrs["placeholder"] != "Name" {
		t.Errorf("placeholder = %q", h.Attrs["placeholder"])
	}
	// Controlled value syncs through OnMount, not the value attribute.
	if h.OnMount == nil {
		t.Error("Input must set OnMount to sync the value property")
	}
	if _, ok := h.Attrs["value"]; ok {
		t.Error("Input must not set the value attribute (caret churn); it uses the property")
	}
}

func TestInputTypeAndConstraints(t *testing.T) {
	h := hostOfCtx(t, Input{Type: InputNumber, Min: "0", Max: "10", Step: "any"}, testCtx(themes.Apple))
	if h.Attrs["type"] != "number" {
		t.Errorf("type = %q, want number", h.Attrs["type"])
	}
	if h.Attrs["min"] != "0" || h.Attrs["max"] != "10" || h.Attrs["step"] != "any" {
		t.Errorf("constraints not mapped: %v", h.Attrs)
	}
}

func TestInputOnChangedFiresWithValue(t *testing.T) {
	var got string
	h := hostOfCtx(t, Input{OnChanged: func(v string) { got = v }}, testCtx(themes.Apple))
	if h.Events["input"] == nil {
		t.Fatal("Input with OnChanged must wire the input event")
	}
	h.Events["input"](gutter.Event{Type: "input", Value: "typed"})
	if got != "typed" {
		t.Errorf("OnChanged received %q, want %q", got, "typed")
	}
}

func TestInputErrorBorder(t *testing.T) {
	th := themes.Apple
	normal := hostOfCtx(t, Input{}, testCtx(th))
	errored := hostOfCtx(t, Input{Error: true}, testCtx(th))
	if normal.Style["border"] == errored.Style["border"] {
		t.Error("Error:true should change the border color")
	}
	if !strings.Contains(errored.Style["border"], th.Components.Input.BorderColorError) {
		t.Errorf("error border = %q, want it to use BorderColorError", errored.Style["border"])
	}
}

func TestCard(t *testing.T) {
	th := themes.Apple
	h := hostOfCtx(t, Card{Variant: CardFeature, Child: Text{Data: "x"}}, testCtx(th))
	wantTag(t, h, "div")
	wantStyle(t, h, "background-color", th.Components.CardFeature.Background)
	wantChildren(t, h, 1)
}

func TestCardPaddingOverride(t *testing.T) {
	h := hostOfCtx(t, Card{Padding: "40px"}, testCtx(themes.Apple))
	wantStyle(t, h, "padding", "40px")
}

func TestIconButton(t *testing.T) {
	clicked := false
	h := hostOfCtx(t, IconButton{Icon: "menu", Tooltip: "Open menu", OnPressed: func() { clicked = true }}, testCtx(themes.Apple))
	wantTag(t, h, "button")
	if h.Attrs["title"] != "Open menu" || h.Attrs["aria-label"] != "Open menu" {
		t.Errorf("tooltip not exposed as title/aria-label: %v", h.Attrs)
	}
	wantChildren(t, h, 1) // the Icon
	h.Events["click"](gutter.Event{})
	if !clicked {
		t.Error("IconButton OnPressed not invoked")
	}
}

func TestScaffoldAppliesThemeAndCanvas(t *testing.T) {
	ctx := testCtx(themes.Apple)
	// Scaffold.Theme should override the ctx theme for the subtree.
	h := hostOfCtx(t, Scaffold{Theme: themes.Meta, Body: Text{Data: "x"}}, ctx)
	if ctx.Theme != themes.Meta {
		t.Error("Scaffold should set ctx.Theme to its Theme")
	}
	wantStyle(t, h, "background-color", themes.Meta.Colors.Canvas)
	wantStyle(t, h, "color", themes.Meta.Colors.Ink)
	wantStyle(t, h, "display", "flex")
	wantStyle(t, h, "flex-direction", "column")
}

func TestScaffoldComposesChrome(t *testing.T) {
	h := hostOfCtx(t, Scaffold{
		AppBar: Text{Data: "bar"},
		Body:   Text{Data: "body"},
		Footer: Text{Data: "foot"},
	}, testCtx(themes.Apple))
	// AppBar + Body wrapper + Footer.
	wantChildren(t, h, 3)
}

func TestScaffoldTitleDoesNotPanicOnHost(t *testing.T) {
	// SetTitle has a host stub; calling Build with a Title must be safe.
	_ = hostOfCtx(t, Scaffold{Title: "Hello", Body: Text{Data: "x"}}, testCtx(themes.Apple))
}

func TestTransformIdentityZeroValue(t *testing.T) {
	h := hostOf(t, Transform{Child: Text{Data: "x"}})
	// Zero value is the identity: no transform, but display:inline-block set.
	wantNoStyle(t, h, "transform")
	wantStyle(t, h, "display", "inline-block")
	wantChildren(t, h, 1)
}

func TestTransformComposition(t *testing.T) {
	h := hostOf(t, Transform{TranslateX: 10, TranslateY: -5, Rotate: 90, Scale: 2})
	tr := h.Style["transform"]
	for _, want := range []string{"translate(10px, -5px)", "rotate(90deg)", "scale(2, 2)"} {
		if !strings.Contains(tr, want) {
			t.Errorf("transform %q missing %q", tr, want)
		}
	}
}

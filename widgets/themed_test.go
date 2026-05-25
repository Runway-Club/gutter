package widgets

import (
	"strings"
	"testing"

	"github.com/Runway-Club/gutter"
	"github.com/Runway-Club/gutter/themes"
)

func TestButtonRendersButtonTag(t *testing.T) {
	clicked := false
	h := hostOfCtx(t, Button{Label: "Go", OnPressed: func() { clicked = true }}, testCtx(themes.Apple))
	wantTag(t, h, "button")
	if h.Text != "Go" {
		t.Errorf("button text = %q, want %q", h.Text, "Go")
	}
	wantStyle(t, h, "cursor", "pointer")
	// The click handler is wired.
	if h.Events["click"] == nil {
		t.Fatal("button has no click handler")
	}
	h.Events["click"](gutter.Event{Type: "click"})
	if !clicked {
		t.Error("OnPressed not invoked by click handler")
	}
}

func TestButtonChildWins(t *testing.T) {
	h := hostOfCtx(t, Button{Label: "ignored", Child: Text{Data: "x"}}, testCtx(themes.Apple))
	if h.Text != "" {
		t.Errorf("Child should suppress Label text, got %q", h.Text)
	}
	wantChildren(t, h, 1)
}

func TestButtonVariantsUseThemeColors(t *testing.T) {
	th := themes.Apple
	primary := hostOfCtx(t, Button{Variant: ButtonPrimary, Label: "x"}, testCtx(th))
	wantStyle(t, primary, "background-color", th.Components.ButtonPrimary.Background)
	ghost := hostOfCtx(t, Button{Variant: ButtonGhost, Label: "x"}, testCtx(th))
	wantStyle(t, ghost, "background-color", th.Components.ButtonGhost.Background)
}

func TestImageAssetResolvesThroughAssetURL(t *testing.T) {
	defer gutter.SetAssetBase("")
	gutter.SetAssetBase("")
	h := hostOf(t, Image{Asset: "logo.png", Alt: "Logo"})
	wantTag(t, h, "img")
	if h.Attrs["src"] != "assets/logo.png" {
		t.Errorf("img src = %q, want %q", h.Attrs["src"], "assets/logo.png")
	}
	if h.Attrs["alt"] != "Logo" {
		t.Errorf("img alt = %q, want %q", h.Attrs["alt"], "Logo")
	}
}

func TestImageSrcUsedDirectly(t *testing.T) {
	h := hostOf(t, Image{Src: "https://cdn/x.png", Fit: ImageFitCover, Rounded: "50%"})
	if h.Attrs["src"] != "https://cdn/x.png" {
		t.Errorf("img src = %q, want absolute URL unchanged", h.Attrs["src"])
	}
	wantStyle(t, h, "object-fit", "cover")
	wantStyle(t, h, "border-radius", "50%")
}

func TestImageAssetWinsOverSrc(t *testing.T) {
	defer gutter.SetAssetBase("")
	gutter.SetAssetBase("")
	h := hostOf(t, Image{Asset: "a.png", Src: "https://cdn/b.png"})
	if h.Attrs["src"] != "assets/a.png" {
		t.Errorf("Asset should win over Src; src = %q", h.Attrs["src"])
	}
}

func TestBadgeVariants(t *testing.T) {
	th := themes.Apple
	h := hostOfCtx(t, Badge{Variant: BadgeSuccess, Text: "In stock"}, testCtx(th))
	wantTag(t, h, "span")
	if h.Text != "In stock" {
		t.Errorf("badge text = %q", h.Text)
	}
	wantStyle(t, h, "background-color", th.Components.BadgeSuccess.Background)
}

func TestIconRendersGlyphAndClass(t *testing.T) {
	h := hostOf(t, Icon{Name: "home"})
	wantTag(t, h, "span")
	if h.Text != "home" {
		t.Errorf("icon glyph text = %q, want %q", h.Text, "home")
	}
	if h.Attrs["class"] != "material-symbols-outlined" {
		t.Errorf("icon class = %q", h.Attrs["class"])
	}
	wantStyle(t, h, "font-size", "24px") // default size
	if !strings.Contains(h.Style["font-variation-settings"], "'FILL' 0") {
		t.Errorf("unfilled icon should have FILL 0, got %q", h.Style["font-variation-settings"])
	}
}

func TestIconFilledAndStyle(t *testing.T) {
	h := hostOf(t, Icon{Name: "star", Filled: true, Style: IconRounded, Size: "32px", Weight: 600})
	if h.Attrs["class"] != "material-symbols-rounded" {
		t.Errorf("icon class = %q, want rounded", h.Attrs["class"])
	}
	wantStyle(t, h, "font-size", "32px")
	v := h.Style["font-variation-settings"]
	if !strings.Contains(v, "'FILL' 1") || !strings.Contains(v, "'wght' 600") {
		t.Errorf("variation settings = %q, want FILL 1 + wght 600", v)
	}
}

func TestHeadingLevelsMapToTypography(t *testing.T) {
	th := themes.Apple
	cases := []struct {
		level HeadingLevel
		spec  themes.TextSpec
	}{
		{H1, th.Typography.HeroDisplay},
		{H2, th.Typography.DisplayLarge},
		{H3, th.Typography.DisplayMedium},
		{H4, th.Typography.HeadingLarge},
		{H5, th.Typography.HeadingMedium},
		{H6, th.Typography.HeadingSmall},
	}
	for _, c := range cases {
		h := hostOfCtx(t, Heading{Level: c.level, Text: "x"}, testCtx(th))
		if h.Style["font-size"] != c.spec.FontSize {
			t.Errorf("H%d font-size = %q, want %q", c.level, h.Style["font-size"], c.spec.FontSize)
		}
	}
}

func TestHeadingRendersSemanticTag(t *testing.T) {
	th := themes.Apple
	cases := map[HeadingLevel]string{H1: "h1", H2: "h2", H3: "h3", H4: "h4", H5: "h5", H6: "h6"}
	for level, tag := range cases {
		h := hostOfCtx(t, Heading{Level: level, Text: "x"}, testCtx(th))
		wantTag(t, h, tag)
		if h.Style["margin"] != "0" {
			t.Errorf("H%d should reset margin to 0, got %q", level, h.Style["margin"])
		}
	}
}

func TestLinkHref(t *testing.T) {
	th := themes.Apple
	real := hostOfCtx(t, Link{Text: "Docs", Href: "/docs"}, testCtx(th))
	if real.Attrs["href"] != "/docs" {
		t.Errorf("href = %q, want /docs", real.Attrs["href"])
	}
	js := hostOfCtx(t, Link{Text: "x", OnPressed: func() {}}, testCtx(th))
	if js.Attrs["href"] != "javascript:void(0)" {
		t.Errorf("JS-driven link href = %q", js.Attrs["href"])
	}
}

func TestBodyVariants(t *testing.T) {
	th := themes.Apple
	strong := hostOfCtx(t, Body{Text: "x", Bold: true}, testCtx(th))
	if strong.Style["font-size"] != th.Typography.BodyStrong.FontSize {
		t.Errorf("Bold body should use BodyStrong spec")
	}
	small := hostOfCtx(t, Body{Text: "x", Small: true}, testCtx(th))
	if small.Style["font-size"] != th.Typography.Caption.FontSize {
		t.Errorf("Small body should use Caption spec")
	}
	// Caption is shorthand for Body{Small:true}.
	cap := hostOfCtx(t, Caption{Text: "x"}, testCtx(th))
	if cap.Style["font-size"] != small.Style["font-size"] {
		t.Errorf("Caption should match Body{Small:true}")
	}
}

func TestActiveThemeFallsBackToApple(t *testing.T) {
	// nil ctx / nil theme must not panic — activeTheme falls back to Apple.
	h := hostOfCtx(t, Body{Text: "x"}, &gutter.BuildContext{})
	wantStyle(t, h, "color", themes.Apple.Colors.Ink)
	h2 := hostOfCtx(t, Body{Text: "x"}, nil)
	wantStyle(t, h2, "color", themes.Apple.Colors.Ink)
}
